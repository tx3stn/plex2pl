// Package jellyfin holds logic related to creating jellyfin native playlist files.
package jellyfin

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/tx3stn/plex2pl/internal/fsname"
	"github.com/tx3stn/plex2pl/internal/plex"
)

const (
	// header is the XML declaration jellyfin writes at the top of its playlist files,
	// which differs from Go's xml.Header.
	header = `<?xml version="1.0" encoding="utf-8" standalone="yes"?>` + "\n"

	// addedTimeFormat is the timestamp format jellyfin uses for the Added element.
	addedTimeFormat = "01/02/2006 15:04:05"

	millisecondsPerMinute = 60_000
)

// Playlist contains the data required to create the jellyfin playlist file.
type Playlist struct {
	XMLName           xml.Name       `xml:"Item"`
	Added             string         `xml:"Added"`
	LockData          bool           `xml:"LockData"`
	LocalTitle        string         `xml:"LocalTitle"`
	RunningTime       int            `xml:"RunningTime"`
	Genres            []string       `xml:"Genres>Genre"`
	OwnerUserID       string         `xml:"OwnerUserId,omitempty"`
	PlaylistItems     []PlaylistItem `xml:"PlaylistItems>PlaylistItem"`
	Shares            struct{}       `xml:"Shares"`
	PlaylistMediaType string         `xml:"PlaylistMediaType"`

	durationMS int
}

// NewPlaylist creates a new instance of the playlist struct.
func NewPlaylist(title string, addedAt int, ownerUserID string) *Playlist {
	return &Playlist{
		Added:             time.Unix(int64(addedAt), 0).UTC().Format(addedTimeFormat),
		LockData:          false,
		LocalTitle:        title,
		Genres:            []string{},
		OwnerUserID:       ownerUserID,
		PlaylistItems:     []PlaylistItem{},
		PlaylistMediaType: "Audio",
	}
}

// AddItem adds an item to the playlist, accumulating the track duration and any
// genres not already included.
func (p *Playlist) AddItem(input plex.PlaylistItem) {
	p.PlaylistItems = append(p.PlaylistItems, PlaylistItem{
		Path: input.Media[0].Part[0].File,
	})

	p.durationMS += input.Media[0].Duration

	for _, genre := range input.Genre {
		if !slices.Contains(p.Genres, genre.Tag) {
			p.Genres = append(p.Genres, genre.Tag)
		}
	}
}

// WriteFile writes the playlist.xml file to a directory named after the playlist
// inside the specified path, matching jellyfin's native playlist layout.
func (p *Playlist) WriteFile(dirPath string) error {
	p.RunningTime = p.durationMS / millisecondsPerMinute

	body, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling playlist xml: %w", err)
	}

	playlistDir := filepath.Join(dirPath, fsname.Sanitize(p.LocalTitle))
	if err := os.MkdirAll(filepath.Clean(playlistDir), 0o750); err != nil {
		return fmt.Errorf("error creating playlist directory: %w", err)
	}

	outPath := filepath.Join(playlistDir, "playlist.xml")

	file, err := os.Create(filepath.Clean(outPath))
	if err != nil {
		return fmt.Errorf("error creating playlist.xml file: %w", err)
	}

	if _, err := file.WriteString(header + string(body) + "\n"); err != nil {
		// The write error takes precedence over any close error.
		_ = file.Close()

		return fmt.Errorf("error writing playlist.xml file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("error closing playlist.xml file: %w", err)
	}

	return nil
}

// PlaylistItem is the data required to create each entry inside the playlist.
type PlaylistItem struct {
	Path string `xml:"Path"`
}
