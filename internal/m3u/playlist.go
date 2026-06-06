// Package m3u hold log related to creating the m3u playlist files.
package m3u

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tx3stn/plex2pl/internal/fsname"
	"github.com/tx3stn/plex2pl/internal/plex"
)

// Playlist contains the data required to create the playlist file.
type Playlist struct {
	Title string
	Items []PlaylistItem
}

// NewPlaylist creates a new instance of thew playlist struct.
func NewPlaylist(title string) *Playlist {
	return &Playlist{
		Title: title,
		Items: []PlaylistItem{},
	}
}

// AddItem adds an item to the playlist.
func (p *Playlist) AddItem(item PlaylistItem) {
	p.Items = append(p.Items, item)
}

// WriteFile writes the m3u file to the specified path.
func (p *Playlist) WriteFile(dirPath string) error {
	output := `#EXTM3U

`

	var outputSb38 strings.Builder
	for _, item := range p.Items {
		outputSb38.WriteString(item.FormatOutput())
	}

	output += outputSb38.String()

	outPath := filepath.Join(dirPath, fsname.Sanitize(p.Title)+".m3u")

	file, err := os.Create(filepath.Clean(outPath))
	if err != nil {
		return fmt.Errorf("error creating m3u file: %w", err)
	}

	if _, err := file.WriteString(output); err != nil {
		// The write error takes precedence over any close error.
		_ = file.Close()

		return fmt.Errorf("error writing m3u file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("error closing m3u file: %w", err)
	}

	return nil
}

// PlaylistItem is the data required to create each entry inside the playlist.
type PlaylistItem struct {
	Duration  int
	FilePath  string
	Artist    string
	TrackName string
	AlbumName string
}

// NewPlaylistItem takes the plex.PlaylistItem data and converts it into the format
// required for the m3u file.
func NewPlaylistItem(input plex.PlaylistItem) PlaylistItem {
	return PlaylistItem{
		Duration:  input.Media[0].Duration,
		FilePath:  input.Media[0].Part[0].File,
		Artist:    input.GrandParentTitle,
		TrackName: input.Title,
		AlbumName: input.ParentTitle,
	}
}

// FormatOutput creates the m3u output string for the playlist item that gets written to file.
func (p PlaylistItem) FormatOutput() string {
	return fmt.Sprintf(`#EXTINF:%d,%s
%s

`, p.Duration, p.TrackName, p.FilePath)
}
