// Package cmd contains the commands that can be run.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/tx3stn/plex2pl/internal/config"
	"github.com/tx3stn/plex2pl/internal/flags"
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

	ctx := context.Background()

	playlists, err := p.GetAudioPlaylists(ctx)
	if err != nil {
		return fmt.Errorf("error getting playlists: %w", err)
	}

	log.Info("found %d audio playlists", len(playlists))

	for i, v := range playlists {
		log.Info("creating playlist %d: '%s'", i, v.Title)

		m3uPlaylist := m3u.NewPlaylist(v.Title)

		items, err := p.GetPlaylistItems(ctx, v.RatingKey)
		if errors.Is(err, plex.ErrNoItemsInPlaylist) {
			log.Info("playlist '%s' contains no items", v.Title)

			continue
		}

		if err != nil {
			return fmt.Errorf("error getting playlist items: %w", err)
		}

		for _, item := range items {
			m3uPlaylist.AddItem(m3u.NewPlaylistItem(item))
		}

		if err = m3uPlaylist.WriteFile(cfg.OutDirectory); err != nil {
			return fmt.Errorf("error writing m3u file: %w", err)
		}
	}

	return nil
}
