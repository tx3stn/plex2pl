// Package cmd contains the commands that can be run.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/tx3stn/plex2pl/internal/config"
	"github.com/tx3stn/plex2pl/internal/flags"
	"github.com/tx3stn/plex2pl/internal/jellyfin"
	"github.com/tx3stn/plex2pl/internal/logger"
	"github.com/tx3stn/plex2pl/internal/m3u"
	"github.com/tx3stn/plex2pl/internal/plex"
)

// Version is the project version set at build time.
//
//nolint:gochecknoglobals
var Version string

// Run runs the CLI.
func Run() error {
	flags.Create()

	log := logger.NewBasic(flags.Verbose)

	cfg, err := config.Get(flags.ConfigFile)
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}

	log.Debug("verbose: %t", flags.Verbose)
	log.Debug("version: %s", Version)
	log.Debug("   file: %s", flags.ConfigFile)
	log.Debug(" config: %+v", cfg)

	client := &http.Client{}
	p := plex.NewClient(client, cfg.PlexServerURL, cfg.PlexAuthToken, log)

	return execute(context.Background(), cfg, p, log)
}

// execute fetches the audio playlists from plex and writes each one to file in the
// configured output format.
func execute(ctx context.Context, cfg config.Config, p *plex.Client, log *logger.Basic) error {
	playlists, err := p.GetAudioPlaylists(ctx)
	if err != nil {
		return fmt.Errorf("error getting playlists: %w", err)
	}

	log.Info("found %d audio playlists", len(playlists))

	// jf holds the state needed to create playlists via the jellyfin API, only
	// populated when using the jellyfin output format.
	var jf *jellyfinExport

	if cfg.OutputFormat == config.FormatJellyfin {
		jf, err = newJellyfinExport(ctx, cfg, log)
		if err != nil {
			return err
		}
	}

	// genreCache stores genres already fetched for a track so tracks appearing in
	// multiple playlists are only queried once.
	genreCache := map[string][]plex.Tag{}

	for i, v := range playlists {
		log.Info("creating playlist %d: '%s'", i, v.Title)

		if err := exportPlaylist(ctx, cfg, p, jf, log, v, genreCache); err != nil {
			return err
		}
	}

	return nil
}

// exportPlaylist writes or creates a single playlist in the configured output format,
// skipping playlists that are empty or, for the jellyfin API format, already exist.
func exportPlaylist(
	ctx context.Context,
	cfg config.Config,
	p *plex.Client,
	jf *jellyfinExport,
	log *logger.Basic,
	playlist plex.Playlist,
	genreCache map[string][]plex.Tag,
) error {
	// For the jellyfin API format, skip playlists that already exist before doing any
	// further work.
	if jf != nil && jf.exists(playlist.Title) {
		log.Info("playlist '%s' already exists in jellyfin, skipping", playlist.Title)

		return nil
	}

	items, err := p.GetPlaylistItems(ctx, playlist.RatingKey)
	if errors.Is(err, plex.ErrNoItemsInPlaylist) {
		log.Info("playlist '%s' contains no items", playlist.Title)

		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting playlist items: %w", err)
	}

	items = playableItems(log, playlist.Title, items)
	if len(items) == 0 {
		log.Info("playlist '%s' contains no playable items", playlist.Title)

		return nil
	}

	switch cfg.OutputFormat {
	case config.FormatM3U:
		return writeM3UPlaylist(cfg, playlist, items)

	case config.FormatJellyfinXML:
		return writeJellyfinPlaylist(ctx, cfg, p, log, playlist, items, genreCache)

	case config.FormatJellyfin:
		return jf.createPlaylist(ctx, log, playlist, items)

	default:
		return fmt.Errorf("%w: %s", config.ErrInvalidOutputFormat, cfg.OutputFormat)
	}
}

// jellyfinExport bundles the state needed to create playlists via the jellyfin API:
// the client, the owner to assign, an index of file path to jellyfin item id, and the
// set of playlist names that already exist.
type jellyfinExport struct {
	client        *jellyfin.Client
	ownerUserID   string
	itemIndex     map[string]string
	existingNames map[string]struct{}
}

// newJellyfinExport builds the jellyfin API client and fetches, once and up front, the
// item index used to resolve plex track paths to jellyfin item ids and the set of
// existing playlist names used to skip playlists that are already present.
func newJellyfinExport(
	ctx context.Context,
	cfg config.Config,
	log *logger.Basic,
) (*jellyfinExport, error) {
	client := jellyfin.NewClient(&http.Client{}, cfg.JellyfinServerURL, cfg.JellyfinAPIKey, log)

	itemIndex, err := client.AudioItemIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building jellyfin item index: %w", err)
	}

	log.Info("indexed %d jellyfin audio items", len(itemIndex))

	existingNames, err := client.PlaylistNames(ctx, cfg.JellyfinOwnerUserID)
	if err != nil {
		return nil, fmt.Errorf("error fetching existing jellyfin playlists: %w", err)
	}

	return &jellyfinExport{
		client:        client,
		ownerUserID:   cfg.JellyfinOwnerUserID,
		itemIndex:     itemIndex,
		existingNames: existingNames,
	}, nil
}

// exists reports whether a playlist with the given name already exists in jellyfin.
func (j *jellyfinExport) exists(name string) bool {
	_, ok := j.existingNames[name]

	return ok
}

// createPlaylist resolves the plex items to jellyfin item ids and creates the playlist
// via the jellyfin API. Tracks with no matching jellyfin item are skipped, and a
// playlist that resolves to no items is skipped rather than created empty.
func (j *jellyfinExport) createPlaylist(
	ctx context.Context,
	log *logger.Basic,
	playlist plex.Playlist,
	items []plex.PlaylistItem,
) error {
	itemIDs := make([]string, 0, len(items))

	for _, item := range items {
		path := item.Media[0].Part[0].File

		id, ok := j.itemIndex[path]
		if !ok {
			log.Info(
				"skipping track '%s' in playlist '%s': no jellyfin item for path %s",
				item.Title,
				playlist.Title,
				path,
			)

			continue
		}

		itemIDs = append(itemIDs, id)
	}

	if len(itemIDs) == 0 {
		log.Info("playlist '%s' has no items that resolve in jellyfin, skipping", playlist.Title)

		return nil
	}

	id, err := j.client.CreatePlaylist(ctx, playlist.Title, itemIDs, j.ownerUserID)
	if err != nil {
		return fmt.Errorf("error creating jellyfin playlist: %w", err)
	}

	log.Info("created jellyfin playlist '%s' (%s) with %d items", playlist.Title, id, len(itemIDs))

	return nil
}

// writeM3UPlaylist writes the playlist items as an m3u file.
func writeM3UPlaylist(cfg config.Config, playlist plex.Playlist, items []plex.PlaylistItem) error {
	out := m3u.NewPlaylist(playlist.Title)

	for _, item := range items {
		out.AddItem(m3u.NewPlaylistItem(item))
	}

	if err := out.WriteFile(cfg.OutDirectory); err != nil {
		return fmt.Errorf("error writing m3u file: %w", err)
	}

	return nil
}

// writeJellyfinPlaylist writes the playlist items as a jellyfin native playlist.xml
// file, resolving the genres for any items that don't include them in the playlist
// items response.
func writeJellyfinPlaylist(
	ctx context.Context,
	cfg config.Config,
	p *plex.Client,
	log *logger.Basic,
	playlist plex.Playlist,
	items []plex.PlaylistItem,
	genreCache map[string][]plex.Tag,
) error {
	if err := resolveGenres(ctx, p, items, genreCache); err != nil {
		// Genres are nice to have, so the playlist is still written without them
		// rather than failing the whole export.
		log.Error("error resolving genres for playlist '%s': %s", playlist.Title, err)
	}

	out := jellyfin.NewPlaylist(playlist.Title, playlist.AddedAt, cfg.JellyfinOwnerUserID)

	for _, item := range items {
		if len(item.Genre) == 0 {
			item.Genre = genreCache[item.RatingKey]
		}

		out.AddItem(item)
	}

	if err := out.WriteFile(cfg.OutDirectory); err != nil {
		return fmt.Errorf("error writing jellyfin playlist file: %w", err)
	}

	return nil
}

// resolveGenres fetches the genres for the tracks that don't include them in the
// playlist items response, using a single batch metadata request for the tracks
// that aren't already in the cache.
func resolveGenres(
	ctx context.Context,
	p *plex.Client,
	items []plex.PlaylistItem,
	genreCache map[string][]plex.Tag,
) error {
	ratingKeys := []string{}

	for _, item := range items {
		if len(item.Genre) > 0 || item.RatingKey == "" {
			continue
		}

		if _, cached := genreCache[item.RatingKey]; cached {
			continue
		}

		if !slices.Contains(ratingKeys, item.RatingKey) {
			ratingKeys = append(ratingKeys, item.RatingKey)
		}
	}

	if len(ratingKeys) == 0 {
		return nil
	}

	tracks, err := p.GetTracksMetadata(ctx, ratingKeys)
	if err != nil {
		return fmt.Errorf("error getting track metadata: %w", err)
	}

	// Cache every requested key, including any plex returned no metadata for, so
	// they aren't requested again for later playlists.
	for _, key := range ratingKeys {
		genreCache[key] = nil
	}

	for _, track := range tracks {
		genreCache[track.RatingKey] = track.Genre
	}

	return nil
}

// playableItems filters out any items that have no file path in the plex response,
// which would otherwise create unplayable entries in the generated playlists.
func playableItems(
	log *logger.Basic,
	playlistTitle string,
	items []plex.PlaylistItem,
) []plex.PlaylistItem {
	playable := make([]plex.PlaylistItem, 0, len(items))

	for _, item := range items {
		if len(item.Media) == 0 || len(item.Media[0].Part) == 0 ||
			item.Media[0].Part[0].File == "" {
			log.Info(
				"skipping item '%s' in playlist '%s': no file in plex response",
				item.Title,
				playlistTitle,
			)

			continue
		}

		playable = append(playable, item)
	}

	return playable
}
