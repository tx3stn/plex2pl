package cmd_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tx3stn/plex2pl/cmd"
	"github.com/tx3stn/plex2pl/internal/config"
	"github.com/tx3stn/plex2pl/internal/logger"
	"github.com/tx3stn/plex2pl/internal/plex"
)

func TestExecuteM3UFormat(t *testing.T) {
	t.Parallel()

	var metadataRequests atomic.Int64

	server := newMockPlexServer(t, batchMetadataHandler(t, &metadataRequests))
	outDir := t.TempDir()

	cfg := config.Config{
		PlexServerURL: server.URL,
		PlexAuthToken: "test-token",
		OutDirectory:  outDir,
		OutputFormat:  config.FormatM3U,
	}

	log := logger.NewBasic(false)
	p := plex.NewClient(&http.Client{}, server.URL, "test-token", log)

	require.NoError(t, cmd.Execute(t.Context(), cfg, p, log))

	assertMatchesExpectedFile(t, filepath.Join(outDir, "2026 jamz.m3u"), "2026 jamz.m3u")
	assertMatchesExpectedFile(
		t,
		filepath.Join(outDir, "repeat offenders.m3u"),
		"repeat offenders.m3u",
	)
	assertMatchesExpectedFile(
		t,
		filepath.Join(outDir, "rock-metal mix.m3u"),
		"rock-metal mix.m3u",
	)
	assert.NoFileExists(t, filepath.Join(outDir, "empty playlist.m3u"))
	assert.NoFileExists(t, filepath.Join(outDir, "movie mondays.m3u"))
	assert.Equal(t, int64(0), metadataRequests.Load(), "m3u format should not query track metadata")
}

func TestExecuteJellyfinFormat(t *testing.T) {
	t.Parallel()

	var metadataRequests atomic.Int64

	server := newMockPlexServer(t, batchMetadataHandler(t, &metadataRequests))
	outDir := t.TempDir()

	cfg := config.Config{
		PlexServerURL:       server.URL,
		PlexAuthToken:       "test-token",
		OutDirectory:        outDir,
		OutputFormat:        config.FormatJellyfin,
		JellyfinOwnerUserID: "0f474ccb9a614c91b69466f2bbb31974",
	}

	log := logger.NewBasic(false)
	p := plex.NewClient(&http.Client{}, server.URL, "test-token", log)

	require.NoError(t, cmd.Execute(t.Context(), cfg, p, log))

	assertMatchesExpectedFile(
		t,
		filepath.Join(outDir, "2026 jamz", "playlist.xml"),
		filepath.Join("2026 jamz", "playlist.xml"),
	)
	assertMatchesExpectedFile(
		t,
		filepath.Join(outDir, "repeat offenders", "playlist.xml"),
		filepath.Join("repeat offenders", "playlist.xml"),
	)
	assertMatchesExpectedFile(
		t,
		filepath.Join(outDir, "rock-metal mix", "playlist.xml"),
		filepath.Join("rock-metal mix", "playlist.xml"),
	)
	assert.NoDirExists(t, filepath.Join(outDir, "empty playlist"))
	assert.NoDirExists(t, filepath.Join(outDir, "movie mondays"))
	assert.Equal(
		t,
		int64(1),
		metadataRequests.Load(),
		"tracks without genres should be queried in a single batch, cached across playlists",
	)
}

func TestExecuteJellyfinFormatWritesPlaylistsWhenMetadataRequestFails(t *testing.T) {
	t.Parallel()

	var metadataRequests atomic.Int64

	server := newMockPlexServer(t, func(w http.ResponseWriter, _ *http.Request) {
		metadataRequests.Add(1)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})
	outDir := t.TempDir()

	cfg := config.Config{
		PlexServerURL:       server.URL,
		PlexAuthToken:       "test-token",
		OutDirectory:        outDir,
		OutputFormat:        config.FormatJellyfin,
		JellyfinOwnerUserID: "0f474ccb9a614c91b69466f2bbb31974",
	}

	log := logger.NewBasic(false)
	p := plex.NewClient(&http.Client{}, server.URL, "test-token", log)

	require.NoError(t, cmd.Execute(t.Context(), cfg, p, log))

	content, err := os.ReadFile(
		filepath.Clean(filepath.Join(outDir, "2026 jamz", "playlist.xml")),
	)
	require.NoError(t, err)

	assert.Contains(t, string(content), "<Genre>Black Metal</Genre>")
	assert.NotContains(t, string(content), "Metalcore")
	assert.Contains(t, string(content), "12 - Jane Doe.mp3")

	assert.FileExists(t, filepath.Join(outDir, "repeat offenders", "playlist.xml"))
	assert.FileExists(t, filepath.Join(outDir, "rock-metal mix", "playlist.xml"))
	assert.Equal(
		t,
		int64(2),
		metadataRequests.Load(),
		"failed lookups should not be cached, so the tracks are retried for the next playlist",
	)
}

// newMockPlexServer creates a test server that responds like a plex server using
// the json stubs in testdata/responses, with the track metadata endpoint served by
// the provided handler.
func newMockPlexServer(t *testing.T, metadataHandler http.HandlerFunc) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	routes := map[string]string{
		"/playlists":          "playlists.json",
		"/playlists/10/items": "playlist_10_items.json",
		"/playlists/11/items": "playlist_11_items.json",
		"/playlists/13/items": "playlist_13_items.json",
		"/playlists/14/items": "playlist_14_items.json",
	}

	for route, stub := range routes {
		mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join("testdata", "responses", stub))
		})
	}

	mux.HandleFunc("/library/metadata/", metadataHandler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server
}

// batchMetadataHandler serves the batch track metadata stub, counting the requests
// made to the endpoint.
func batchMetadataHandler(t *testing.T, metadataRequests *atomic.Int64) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		metadataRequests.Add(1)

		assert.Equal(t, "/library/metadata/201,203", r.URL.Path)
		http.ServeFile(
			w,
			r,
			filepath.Join("testdata", "responses", "library_metadata_201_203.json"),
		)
	}
}

// assertMatchesExpectedFile compares a generated file against the expected output
// file in testdata.
func assertMatchesExpectedFile(t *testing.T, actualPath string, expectedName string) {
	t.Helper()

	actual, err := os.ReadFile(filepath.Clean(actualPath))
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Clean(filepath.Join("testdata", expectedName)))
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(actual))
}
