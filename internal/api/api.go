// Package api contains functions useful for making API requests.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tx3stn/plex2pl/internal/logger"
)

// HTTPClient is a convenience interface to make testing FetchJSON easier.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// FetchJSON makes a GET request and returns the json response marshalled as type T.
//
//nolint:ireturn
func FetchJSON[T any](
	ctx context.Context,
	client HTTPClient,
	url string,
	log *logger.Basic,
) (T, error) {
	return SendJSON[T](ctx, client, http.MethodGet, url, nil, nil, log)
}

// SendJSON makes a request with the given method, optionally sending body as a JSON
// payload and applying the provided headers, returning the json response marshalled
// as type T.
//
//nolint:ireturn
func SendJSON[T any](
	ctx context.Context,
	client HTTPClient,
	method string,
	url string,
	body any,
	headers map[string]string,
	log *logger.Basic,
) (T, error) {
	var responseJSON T

	req, err := buildJSONRequest(ctx, method, url, body, headers)
	if err != nil {
		return responseJSON, err
	}

	log.Debug("%s url: %s", method, url)

	resp, err := client.Do(req)
	if err != nil {
		log.Error(ErrMakingRequest.Error(), err)

		return responseJSON, fmt.Errorf("%w: %w", ErrMakingRequest, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal("error closing response body: %s", err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("error reading response body: %s", err)

		return responseJSON, fmt.Errorf("error reading response body: %w", err)
	}

	log.Debug("response received: %s", respBody)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		log.Error("%s: %d", ErrUnexpectedStatus.Error(), resp.StatusCode)

		return responseJSON, fmt.Errorf(
			"%w: %d: %s",
			ErrUnexpectedStatus,
			resp.StatusCode,
			respBody,
		)
	}

	// A successful response with no body (e.g. a 204) leaves the zero value of T.
	if len(respBody) == 0 {
		return responseJSON, nil
	}

	if err := json.Unmarshal(respBody, &responseJSON); err != nil {
		log.Error(ErrMashallingJSONToType.Error(), err)

		return responseJSON, fmt.Errorf("%w: %w", ErrMashallingJSONToType, err)
	}

	return responseJSON, nil
}

// buildJSONRequest builds an http request, encoding body as a JSON payload when set
// and applying the given headers.
func buildJSONRequest(
	ctx context.Context,
	method string,
	url string,
	body any,
	headers map[string]string,
) (*http.Request, error) {
	var reqBody io.Reader

	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshalling request body: %w", err)
		}

		reqBody = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}

	req.Header.Add("Accept", "application/json")

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	return req, nil
}
