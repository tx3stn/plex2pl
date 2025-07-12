// Package flags contains logic to do with the global variables used as CLI flags.
package flags

import "flag"

var (
	// ConfigFile is the variable for the `--config` CLI flag.
	ConfigFile string

	// Verbose is the variable for the `--verbose` CLI flag.
	Verbose bool
)

// Create creates the flag variables then parses them so set the values at run time.
func Create() {
	flag.StringVar(&ConfigFile, "config", "", "config file path if not in default location")
	flag.BoolVar(&Verbose, "verbose", false, "run in verbose mode")

	flag.Parse()
}
