// Package cmd contains the commands that can be run.
package cmd

import (
	"fmt"

	"github.com/tx3stn/plex2m3u/internal/config"
	"github.com/tx3stn/plex2m3u/internal/flags"
	"github.com/tx3stn/plex2m3u/internal/logger"
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

	return nil
}
