// Package api contains functions useful for making API requests.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tx3stn/plex2m3u/internal/logger"
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
	var responseJSON T

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return responseJSON, fmt.Errorf("error building request: %w", err)
	}

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("error reading response body: %s", err)

		return responseJSON, fmt.Errorf("error reading response body: %w", err)
	}

	log.Debug("response received: %s", body)

	if err := json.Unmarshal(body, &responseJSON); err != nil {
		log.Error(ErrMashallingJSONToType.Error(), err)

		return responseJSON, fmt.Errorf("%w: %w", ErrMashallingJSONToType, err)
	}

	return responseJSON, nil
}
