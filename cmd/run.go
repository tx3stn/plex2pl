// Package cmd contains the commands that can be run.
package cmd

import (
	"context"
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

	for i, v := range playlists {
		log.Info("%d: %+v", i, v)
	}

	// TODO:
	// 1. filter out to audio only
	// 2. get playlist by ID to get track data
	// 3. add track data to m3u file
	// 4. save m3u playlist

	return nil
}
