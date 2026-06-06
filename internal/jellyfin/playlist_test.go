package jellyfin_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tx3stn/plex2pl/internal/jellyfin"
	"github.com/tx3stn/plex2pl/internal/plex"
)

func TestWriteFile(t *testing.T) {
	t.Parallel()

	// 2026-05-16T11:25:58Z
	addedAt := 1778930758

	playlist := jellyfin.NewPlaylist("2026 jamz", addedAt, "0f474ccb9a614c91b69466f2bbb31974")

	playlist.AddItem(plex.PlaylistItem{
		Title: "Bodhidharma",
		Genre: []plex.Tag{{Tag: "Black Metal"}, {Tag: "Ecstatic Black Metal"}},
		Media: []plex.PlaylistMedia{
			{
				Duration: 420_000,
				Part: []plex.MediaPart{
					{File: "/music/Agriculture/The Spiritual Sound (2025)/02 - Bodhidharma.flac"},
				},
			},
		},
	})

	playlist.AddItem(plex.PlaylistItem{
		Title: "Jane Doe",
		Genre: []plex.Tag{{Tag: "Metalcore"}, {Tag: "Black Metal"}},
		Media: []plex.PlaylistMedia{
			{
				Duration: 690_000,
				Part: []plex.MediaPart{
					{
						File: "/music/Converge/Jane Doe (2001)/12 - Jane Doe.mp3",
					},
				},
			},
		},
	})

	dir := t.TempDir()

	require.NoError(t, playlist.WriteFile(dir))

	content, err := os.ReadFile(filepath.Clean(filepath.Join(dir, "2026 jamz", "playlist.xml")))
	require.NoError(t, err)

	expected := `<?xml version="1.0" encoding="utf-8" standalone="yes"?>
<Item>
  <Added>05/16/2026 11:25:58</Added>
  <LockData>false</LockData>
  <LocalTitle>2026 jamz</LocalTitle>
  <RunningTime>18</RunningTime>
  <Genres>
    <Genre>Black Metal</Genre>
    <Genre>Ecstatic Black Metal</Genre>
    <Genre>Metalcore</Genre>
  </Genres>
  <OwnerUserId>0f474ccb9a614c91b69466f2bbb31974</OwnerUserId>
  <PlaylistItems>
    <PlaylistItem>
      <Path>/music/Agriculture/The Spiritual Sound (2025)/02 - Bodhidharma.flac</Path>
    </PlaylistItem>
    <PlaylistItem>
      <Path>/music/Converge/Jane Doe (2001)/12 - Jane Doe.mp3</Path>
    </PlaylistItem>
  </PlaylistItems>
  <Shares></Shares>
  <PlaylistMediaType>Audio</PlaylistMediaType>
</Item>
`

	assert.Equal(t, expected, string(content))
}

func TestWriteFileSanitizesPlaylistTitleDirectory(t *testing.T) {
	t.Parallel()

	playlist := jellyfin.NewPlaylist("rock/metal mix", 1778930758, "")

	playlist.AddItem(plex.PlaylistItem{
		Title: "Slaughterhouse",
		Media: []plex.PlaylistMedia{
			{
				Duration: 192_000,
				Part: []plex.MediaPart{
					{File: "/music/Chat Pile/God's Country (2022)/01 - Slaughterhouse.mp3"},
				},
			},
		},
	})

	dir := t.TempDir()

	require.NoError(t, playlist.WriteFile(dir))

	content, err := os.ReadFile(
		filepath.Clean(filepath.Join(dir, "rock-metal mix", "playlist.xml")),
	)
	require.NoError(t, err)

	assert.Contains(
		t,
		string(content),
		"<LocalTitle>rock/metal mix</LocalTitle>",
		"the playlist title should keep the original name, only the directory is sanitized",
	)
}

func TestWriteFileWithoutOwnerOrGenres(t *testing.T) {
	t.Parallel()

	playlist := jellyfin.NewPlaylist("no frills", 1778930758, "")

	playlist.AddItem(plex.PlaylistItem{
		Title: "My Garden",
		Media: []plex.PlaylistMedia{
			{
				Duration: 200_000,
				Part: []plex.MediaPart{
					{File: "/music/Agriculture/The Spiritual Sound (2025)/04 - My Garden.flac"},
				},
			},
		},
	})

	dir := t.TempDir()

	require.NoError(t, playlist.WriteFile(dir))

	content, err := os.ReadFile(filepath.Clean(filepath.Join(dir, "no frills", "playlist.xml")))
	require.NoError(t, err)

	expected := `<?xml version="1.0" encoding="utf-8" standalone="yes"?>
<Item>
  <Added>05/16/2026 11:25:58</Added>
  <LockData>false</LockData>
  <LocalTitle>no frills</LocalTitle>
  <RunningTime>3</RunningTime>
  <Genres></Genres>
  <PlaylistItems>
    <PlaylistItem>
      <Path>/music/Agriculture/The Spiritual Sound (2025)/04 - My Garden.flac</Path>
    </PlaylistItem>
  </PlaylistItems>
  <Shares></Shares>
  <PlaylistMediaType>Audio</PlaylistMediaType>
</Item>
`

	assert.Equal(t, expected, string(content))
}
