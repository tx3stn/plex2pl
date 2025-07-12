// Package cmd contains the commands that can be run.
package cmd

import (
	"fmt"

	"github.com/tx3stn/plex2m3u/internal/flags"
)

// Version is the project version set at build time.
//
//nolint:gochecknoglobals
var Version string

// Run runs the CLI.
func Run() error {
	flags.Create()
	fmt.Printf("cfg: %s\n", flags.ConfigFile)
	fmt.Printf("verbose: %t\n", flags.Verbose)
	fmt.Printf("version: %s\n", Version)

	return nil
}
