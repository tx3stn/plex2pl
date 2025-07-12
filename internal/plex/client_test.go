package plex_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tx3stn/plex2m3u/internal/api"
	"github.com/tx3stn/plex2m3u/internal/api/apitest"
	"github.com/tx3stn/plex2m3u/internal/logger"
	"github.com/tx3stn/plex2m3u/internal/plex"
)

const (
	mockURL     = "http://testing.notreal"
	noPlaylists = `{"mediaContainer":{"size": 0,"metadata":[]}}`
	playlists   = `{
	"mediaContainer":{
		"size":0,
		"metadata":[
			{"title":"2020 jamz","playlistType":"audio"},
			{"title":"2025 jamz","playlistType":"audio"},
			{"title":"movie mondays","playlistType":"video"}
		]
	}
}`
	onlyVideoPlaylists = `{
	"mediaContainer":{
		"size":0,
		"metadata":[
			{"title":"movie mondays","playlistType":"video"},
			{"title":"top files of 2025","playlistType":"video"}
		]
	}
}`
)

func TestGetPlaylists(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		httpClient    func() api.HTTPClient
		expectedError error
		expected      []plex.Playlist
	}{
		"ReturnsErrorIfNoPlaylistsInResponse": {
			httpClient:    mockHTTPClientWithResponse(t, noPlaylists),
			expectedError: plex.ErrNoPlaylists,
			expected:      []plex.Playlist{},
		},
		"ReturnsPlaylistsIncludedInReponse": {
			httpClient:    mockHTTPClientWithResponse(t, playlists),
			expectedError: nil,
			expected: []plex.Playlist{
				{Title: "2020 jamz", PlaylistType: "audio"},
				{Title: "2025 jamz", PlaylistType: "audio"},
				{Title: "movie mondays", PlaylistType: "video"},
			},
		},
	}

	for name, testCase := range testCases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pc := plex.NewClient(
				tc.httpClient(),
				mockURL,
				"fooToken",
				logger.NewBasic(false),
			)

			actual, err := pc.GetPlaylists(t.Context())
			require.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetAudioPlaylists(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		httpClient    func() api.HTTPClient
		expectedError error
		expected      []plex.Playlist
	}{
		"ReturnsOnlyAudioPlaylistsIncludedInReponse": {
			httpClient:    mockHTTPClientWithResponse(t, playlists),
			expectedError: nil,
			expected: []plex.Playlist{
				{Title: "2020 jamz", PlaylistType: "audio"},
				{Title: "2025 jamz", PlaylistType: "audio"},
			},
		},
		"ReturnsErrorIfNoAudioPlaylistsInResponse": {
			httpClient:    mockHTTPClientWithResponse(t, onlyVideoPlaylists),
			expectedError: plex.ErrNoAudioPlaylists,
			expected:      []plex.Playlist{},
		},
	}

	for name, testCase := range testCases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pc := plex.NewClient(
				tc.httpClient(),
				mockURL,
				"fooToken",
				logger.NewBasic(false),
			)

			actual, err := pc.GetAudioPlaylists(t.Context())
			require.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func mockHTTPClientWithResponse(t *testing.T, responseJSON string) func() api.HTTPClient {
	t.Helper()

	return func() api.HTTPClient {
		client := apitest.NewMockHTTPClient(t)

		client.EXPECT().
			Do(mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == http.MethodGet &&
					strings.Contains(req.URL.String(), mockURL+"/playlists?X-Plex-Token=")
			})).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(responseJSON))),
			}, nil)

		return client
	}
}
