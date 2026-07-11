// Package config contains logic for getting configuration options.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tx3stn/plex2pl/internal/flags"
)

const (
	// FormatM3U is the output format value for m3u playlist files.
	FormatM3U = "m3u"
	// FormatJellyfin is the output format value for creating playlists directly in
	// jellyfin via its API, which produces editable, owned playlists.
	FormatJellyfin = "jellyfin"
	// FormatJellyfinXML is the output format value for writing jellyfin native
	// playlist.xml files to disk. These are read-only once jellyfin imports them.
	FormatJellyfinXML = "jellyfin-xml"
)

// Config represents the configuration options required to be defined in the config file.
type Config struct {
	PlexServerURL       string `json:"plexServerUrl"`
	PlexAuthToken       string `json:"plexAuthToken"`
	OutDirectory        string `json:"outDirectory"`
	OutputFormat        string `json:"outputFormat"`
	JellyfinServerURL   string `json:"jellyfinServerUrl"`
	JellyfinAPIKey      string `json:"jellyfinApiKey"`
	JellyfinOwnerUserID string `json:"jellyfinOwnerUserId"`
}

// Get returns the config read from the file.
func Get(fileFlag string) (Config, error) {
	var file string

	var err error

	if fileFlag != "" {
		_, err := os.Stat(fileFlag)
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("%w: %s", ErrConfigNotFound, fileFlag)
		}

		if err != nil {
			return Config{}, fmt.Errorf("error checking for existence of config file: %w", err)
		}

		file = fileFlag
	} else {
		file = FindConfigFile()
		flags.ConfigFile = file
	}

	if file == "" {
		return Config{}, ErrConfigNotFound
	}

	content, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrReadingConfigFile, err)
	}

	var conf Config
	if err = json.Unmarshal(content, &conf); err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrUnmarshalingConfig, err)
	}

	if conf.OutputFormat == "" {
		return Config{}, ErrMissingOutputFormat
	}

	switch conf.OutputFormat {
	case FormatM3U, FormatJellyfinXML:
		// no additional config required for these formats.
	case FormatJellyfin:
		if err := conf.validateJellyfinAPI(); err != nil {
			return Config{}, err
		}
	default:
		return Config{}, fmt.Errorf("%w: %s", ErrInvalidOutputFormat, conf.OutputFormat)
	}

	return conf, nil
}

// FindConfigFile checks the expected paths for a pkb config file and returns the
// path to it if found.
// The paths are checked in the order of precedence:
//   - XDG_CONFIG_DIR
//   - HOME/.config
func FindConfigFile() string {
	paths := []string{}
	dirName := "plex2pl"
	configFileName := "config.json"

	if xdg, ok := os.LookupEnv("XDG_CONFIG_DIR"); ok {
		paths = append(paths, filepath.Join(xdg, dirName))
	}

	if home, ok := os.LookupEnv("HOME"); ok {
		paths = append(paths, filepath.Join(home, ".config", dirName))
	}

	if len(paths) == 0 {
		return ""
	}

	for _, path := range paths {
		file := filepath.Join(path, configFileName)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			// no config file at location, continue looking.
			continue
		}

		return file
	}

	return ""
}

// validateJellyfinAPI ensures the config required to create playlists via the jellyfin
// API is present.
func (c Config) validateJellyfinAPI() error {
	switch {
	case c.JellyfinServerURL == "":
		return ErrMissingJellyfinServerURL
	case c.JellyfinAPIKey == "":
		return ErrMissingJellyfinAPIKey
	case c.JellyfinOwnerUserID == "":
		return ErrMissingJellyfinOwnerUserID
	default:
		return nil
	}
}
