package api_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tx3stn/plex2pl/internal/api"
	"github.com/tx3stn/plex2pl/internal/api/apitest"
	"github.com/tx3stn/plex2pl/internal/logger"
)

const (
	mockURL = "http://theresponseismockedanyway.com"
)

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
			expectedResponse: map[string]any(nil),
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
			expectedResponse: map[string]any(nil),
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
				logger.NewBasic(false),
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

func TestSendJSON(t *testing.T) {
	t.Parallel()

	t.Run("SendsBodyAndHeadersAndReturnsJSON", func(t *testing.T) {
		t.Parallel()

		var captured *http.Request

		client := apitest.NewMockHTTPClient(t)
		client.EXPECT().
			Do(mock.Anything).
			Run(func(req *http.Request) { captured = req }).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"response":"json"}`))),
			}, nil).
			Once()

		response, err := api.SendJSON[map[string]any](
			t.Context(),
			client,
			http.MethodPost,
			mockURL,
			map[string]string{"hello": "world"},
			map[string]string{"Authorization": "token"},
			logger.NewBasic(false),
		)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"response": "json"}, response)

		require.NotNil(t, captured)
		assert.Equal(t, http.MethodPost, captured.Method)
		assert.Equal(t, "application/json", captured.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", captured.Header.Get("Accept"))
		assert.Equal(t, "token", captured.Header.Get("Authorization"))

		body, err := io.ReadAll(captured.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"hello":"world"}`, string(body))
	})

	t.Run("ReturnsErrorOnNon2xxStatus", func(t *testing.T) {
		t.Parallel()

		client := apitest.NewMockHTTPClient(t)
		client.EXPECT().
			Do(mock.Anything).
			Return(&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewReader([]byte(`boom`))),
			}, nil).
			Once()

		_, err := api.SendJSON[map[string]any](
			t.Context(),
			client,
			http.MethodGet,
			mockURL,
			nil,
			nil,
			logger.NewBasic(false),
		)
		require.ErrorIs(t, err, api.ErrUnexpectedStatus)
	})
}
