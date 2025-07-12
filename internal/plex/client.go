// Package plex contains logic related to querying plex.
package plex

import (
	"context"
	"fmt"

	"github.com/tx3stn/plex2m3u/internal/api"
	"github.com/tx3stn/plex2m3u/internal/logger"
)

// Client is the plex client for making requests to yoru plex server.
type Client struct {
	httpClient api.HTTPClient
	serverURL  string
	authToken  string
	log        *logger.Basic
	// playlistFetcher func(context.Context, api.HTTPClient, string, *logger.Basic) (getPlaylistsResponse, error)
}

// NewClient creates a new instance of the Client struct.
func NewClient(httpClient api.HTTPClient, url string, token string, log *logger.Basic) *Client {
	return &Client{
		httpClient: httpClient,
		serverURL:  url,
		authToken:  token,
		log:        log,
		// playlistFetcher: api.FetchJSON[getPlaylistsResponse],
	}
}

// GetPlaylists returns all playlists.
func (c Client) GetPlaylists(ctx context.Context) ([]Playlist, error) {
	playlistsURL := fmt.Sprintf("%s/playlists?X-Plex-Token=%s", c.serverURL, c.authToken)

	c.log.Info("fetching playlists")
	c.log.Debug("playlist url: %s", playlistsURL)

	resp, err := api.FetchJSON[getPlaylistsResponse](ctx, c.httpClient, playlistsURL, c.log)
	if err != nil {
		return nil, err
	}

	c.log.Debug("response: %+v", resp)

	if len(resp.MediaContainer.Metadata) == 0 {
		return []Playlist{}, ErrNoPlaylists
	}

	return resp.MediaContainer.Metadata, nil
}

// getPlaylistsResponse is the response wrapper for the GetPlaylists request to align
// with how plex returns the data.
type getPlaylistsResponse struct {
	MediaContainer MediaContainer `json:"mediaContainer"`
}
