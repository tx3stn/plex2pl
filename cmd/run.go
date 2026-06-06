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

	// genreCache stores genres already fetched for a track so tracks appearing in
	// multiple playlists are only queried once.
	genreCache := map[string][]plex.Tag{}

	for i, v := range playlists {
		log.Info("creating playlist %d: '%s'", i, v.Title)

		items, err := p.GetPlaylistItems(ctx, v.RatingKey)
		if errors.Is(err, plex.ErrNoItemsInPlaylist) {
			log.Info("playlist '%s' contains no items", v.Title)

			continue
		}

		if err != nil {
			return fmt.Errorf("error getting playlist items: %w", err)
		}

		items = playableItems(log, v.Title, items)
		if len(items) == 0 {
			log.Info("playlist '%s' contains no playable items", v.Title)

			continue
		}

		switch cfg.OutputFormat {
		case config.FormatM3U:
			err = writeM3UPlaylist(cfg, v, items)

		case config.FormatJellyfin:
			err = writeJellyfinPlaylist(ctx, cfg, p, log, v, items, genreCache)

		default:
			return fmt.Errorf("%w: %s", config.ErrInvalidOutputFormat, cfg.OutputFormat)
		}

		if err != nil {
			return err
		}
	}

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
