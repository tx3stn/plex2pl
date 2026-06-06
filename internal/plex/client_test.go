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
	"github.com/tx3stn/plex2pl/internal/api"
	"github.com/tx3stn/plex2pl/internal/api/apitest"
	"github.com/tx3stn/plex2pl/internal/logger"
	"github.com/tx3stn/plex2pl/internal/plex"
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
			{"title":"top films of 2025","playlistType":"video"}
		]
	}
}`
	noTrackMetadata = `{"mediaContainer":{"size":0,"metadata":[]}}`
	trackMetadata   = `{
	"mediaContainer":{
		"size":2,
		"metadata":[
			{
				"ratingKey":"101",
				"title":"Bodhidharma",
				"parentTitle":"The Spiritual Sound",
				"grandParentTitle":"Agriculture",
				"genre":[{"tag":"Black Metal"},{"tag":"Ecstatic Black Metal"}]
			},
			{
				"ratingKey":"102",
				"title":"The Saddest Day",
				"parentTitle":"Petitioning the Empty Sky",
				"grandParentTitle":"Converge",
				"genre":[{"tag":"Hardcore Punk"}]
			}
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

func TestGetTracksMetadata(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		httpClient    func() api.HTTPClient
		expectedError error
		expected      []plex.PlaylistItem
	}{
		"ReturnsEmptyWhenNoMetadataInResponse": {
			httpClient:    mockHTTPClientForPath(t, "library/metadata/101,102", noTrackMetadata),
			expectedError: nil,
			expected:      []plex.PlaylistItem{},
		},
		"ReturnsTrackMetadataIncludedInResponse": {
			httpClient:    mockHTTPClientForPath(t, "library/metadata/101,102", trackMetadata),
			expectedError: nil,
			expected: []plex.PlaylistItem{
				{
					RatingKey:        "101",
					Title:            "Bodhidharma",
					ParentTitle:      "The Spiritual Sound",
					GrandParentTitle: "Agriculture",
					Genre: []plex.Tag{
						{Tag: "Black Metal"},
						{Tag: "Ecstatic Black Metal"},
					},
				},
				{
					RatingKey:        "102",
					Title:            "The Saddest Day",
					ParentTitle:      "Petitioning the Empty Sky",
					GrandParentTitle: "Converge",
					Genre: []plex.Tag{
						{Tag: "Hardcore Punk"},
					},
				},
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

			actual, err := pc.GetTracksMetadata(t.Context(), []string{"101", "102"})
			require.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func mockHTTPClientWithResponse(t *testing.T, responseJSON string) func() api.HTTPClient {
	t.Helper()

	return mockHTTPClientForPath(t, "playlists", responseJSON)
}

func mockHTTPClientForPath(t *testing.T, path string, responseJSON string) func() api.HTTPClient {
	t.Helper()

	return func() api.HTTPClient {
		client := apitest.NewMockHTTPClient(t)

		client.EXPECT().
			Do(mock.MatchedBy(func(req *http.Request) bool {
				return req.Method == http.MethodGet &&
					strings.Contains(req.URL.String(), mockURL+"/"+path+"?X-Plex-Token=")
			})).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(responseJSON))),
			}, nil)

		return client
	}
}
