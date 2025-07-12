// Package cmd contains the commands that can be run.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/tx3stn/plex2m3u/internal/config"
	"github.com/tx3stn/plex2m3u/internal/flags"
	"github.com/tx3stn/plex2m3u/internal/logger"
	"github.com/tx3stn/plex2m3u/internal/plex"
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
		log.Info("%d: playlist '%s'", i, v.Title)

		items, err := p.GetPlaylistItems(ctx, v.RatingKey)
		if errors.Is(err, plex.ErrNoItemsInPlaylist) {
			log.Info("playlist '%s' contains no items", v.Title)

			continue
		}

		if err != nil {
			return fmt.Errorf("error getting playlist items: %w", err)
		}

		log.Info("title: %s", items[0].Title)
	}

	// TODO:
	// 3. add track data to m3u file
	// 4. save m3u playlist

	return nil
}
