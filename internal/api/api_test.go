package api_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tx3stn/plex2m3u/internal/api"
	"github.com/tx3stn/plex2m3u/internal/api/apitest"
	"github.com/tx3stn/plex2m3u/internal/logger"
)

const (
	mockURL = "http://theresponseismockedanyway.com"
)

var mockLogger = logger.NewBasic(false)

func TestFetchJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		client           func() api.HTTPClient
		expectedResponse map[string]any
		expectedError    error
	}{
		"ReturnsJSONOnSuccessfulRequest": {
			client: func() api.HTTPClient {
				client := apitest.NewMockHTTPClient(t)
				client.EXPECT().
					Do(expectedRequest(t)).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte(`{"response":"json"}`))),
					}, nil).
					Once()

				return client
			},
			expectedResponse: map[string]any{"response": "json"},
			expectedError:    nil,
		},
		"ReturnsErrorWhenDoRequestFails": {
			client: func() api.HTTPClient {
				client := apitest.NewMockHTTPClient(t)
				client.EXPECT().
					Do(expectedRequest(t)).
					Return(&http.Response{}, errors.New("forced error")).
					Once()

				return client
			},
			expectedResponse: map[string]interface{}(nil),
			expectedError:    api.ErrMakingRequest,
		},
		"ReturnsErrorWhenJSONFailsToUnMarshalToType": {
			client: func() api.HTTPClient {
				client := apitest.NewMockHTTPClient(t)
				client.EXPECT().
					Do(expectedRequest(t)).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte("thishoulderror"))),
					}, nil).
					Once()

				return client
			},
			expectedResponse: map[string]interface{}(nil),
			expectedError:    api.ErrMashallingJSONToType,
		},
	}

	for name, testCase := range testCases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response, err := api.FetchJSON[map[string]any](
				t.Context(),
				tc.client(),
				mockURL,
				mockLogger,
			)
			require.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expectedResponse, response)
		})
	}
}

func expectedRequest(t *testing.T) *http.Request {
	t.Helper()

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, mockURL, nil)
	req.Header.Add("Accept", "application/json")

	return req
}
