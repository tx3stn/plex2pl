package jellyfin_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tx3stn/plex2pl/internal/api/apitest"
	"github.com/tx3stn/plex2pl/internal/jellyfin"
	"github.com/tx3stn/plex2pl/internal/logger"
)

const (
	mockServerURL = "http://jellyfin.local"
	mockAPIKey    = "test-key"
	mockAuth      = `MediaBrowser Token="test-key"`
)

func TestAudioItemIndex(t *testing.T) {
	t.Parallel()

	const items = `{"Items":[
		{"Id":"abc","Name":"Bodhidharma","Path":"/music/a.flac"},
		{"Id":"def","Name":"Jane Doe","Path":"/music/b.mp3"},
		{"Id":"","Name":"missing id","Path":"/music/c.mp3"},
		{"Id":"ghi","Name":"missing path","Path":""}
	]}`

	client := apitest.NewMockHTTPClient(t)
	client.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == http.MethodGet &&
				strings.HasSuffix(req.URL.Path, "/Items") &&
				req.URL.Query().Get("IncludeItemTypes") == "Audio" &&
				req.Header.Get("Authorization") == mockAuth
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(items)),
		}, nil).
		Once()

	jf := jellyfin.NewClient(client, mockServerURL, mockAPIKey, logger.NewBasic(false))

	index, err := jf.AudioItemIndex(t.Context())
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"/music/a.flac": "abc",
		"/music/b.mp3":  "def",
	}, index)
}

func TestPlaylistNames(t *testing.T) {
	t.Parallel()

	const playlists = `{"Items":[
		{"Id":"p1","Name":"2020 jamz"},
		{"Id":"p2","Name":"2025 jamz"},
		{"Id":"p3","Name":""}
	]}`

	client := apitest.NewMockHTTPClient(t)
	client.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == http.MethodGet &&
				strings.HasSuffix(req.URL.Path, "/Items") &&
				req.URL.Query().Get("IncludeItemTypes") == "Playlist" &&
				req.URL.Query().Get("UserId") == "owner-123" &&
				req.Header.Get("Authorization") == mockAuth
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(playlists)),
		}, nil).
		Once()

	jf := jellyfin.NewClient(client, mockServerURL, mockAPIKey, logger.NewBasic(false))

	names, err := jf.PlaylistNames(t.Context(), "owner-123")
	require.NoError(t, err)
	assert.Equal(t, map[string]struct{}{
		"2020 jamz": {},
		"2025 jamz": {},
	}, names)
}

func TestCreatePlaylist(t *testing.T) {
	t.Parallel()

	var captured *http.Request

	client := apitest.NewMockHTTPClient(t)
	client.EXPECT().
		Do(mock.Anything).
		Run(func(req *http.Request) { captured = req }).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"Id":"new-playlist-id"}`)),
		}, nil).
		Once()

	jf := jellyfin.NewClient(client, mockServerURL, mockAPIKey, logger.NewBasic(false))

	id, err := jf.CreatePlaylist(
		t.Context(),
		"2026 jamz",
		[]string{"abc", "def"},
		"owner-123",
	)
	require.NoError(t, err)
	assert.Equal(t, "new-playlist-id", id)

	require.NotNil(t, captured)
	assert.Equal(t, http.MethodPost, captured.Method)
	assert.True(t, strings.HasSuffix(captured.URL.Path, "/Playlists"))
	assert.Equal(t, mockAuth, captured.Header.Get("Authorization"))
	assert.Equal(t, "application/json", captured.Header.Get("Content-Type"))

	body, err := io.ReadAll(captured.Body)
	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{"Name":"2026 jamz","Ids":["abc","def"],"UserId":"owner-123","MediaType":"Audio"}`,
		string(body),
	)
}

func TestCreatePlaylistReturnsErrorOnBadStatus(t *testing.T) {
	t.Parallel()

	client := apitest.NewMockHTTPClient(t)
	client.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(strings.NewReader(`unauthorized`)),
		}, nil).
		Once()

	jf := jellyfin.NewClient(client, mockServerURL, mockAPIKey, logger.NewBasic(false))

	_, err := jf.CreatePlaylist(t.Context(), "2026 jamz", []string{"abc"}, "owner-123")
	require.Error(t, err)
}
