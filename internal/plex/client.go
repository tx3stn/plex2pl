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
}

// NewClient creates a new instance of the Client struct.
func NewClient(httpClient api.HTTPClient, url string, token string, log *logger.Basic) *Client {
	return &Client{
		httpClient: httpClient,
		serverURL:  url,
		authToken:  token,
		log:        log,
	}
}

// GetPlaylists returns all playlists.
func (c Client) GetPlaylists(ctx context.Context) ([]Playlist, error) {
	resp, err := api.FetchJSON[getPlaylistsResponse](ctx, c.httpClient, c.url("playlists"), c.log)
	if err != nil {
		return nil, err
	}

	if len(resp.MediaContainer.Metadata) == 0 {
		return []Playlist{}, ErrNoPlaylists
	}

	return resp.MediaContainer.Metadata, nil
}

// GetAudioPlaylists gets all playlists and filters the response to just audio ones.
func (c Client) GetAudioPlaylists(ctx context.Context) ([]Playlist, error) {
	allPlaylists, err := c.GetPlaylists(ctx)
	if err != nil {
		return nil, err
	}

	audioPlaylists := []Playlist{}

	for _, playlist := range allPlaylists {
		if playlist.PlaylistType == "audio" {
			audioPlaylists = append(audioPlaylists, playlist)
		}
	}

	if len(audioPlaylists) == 0 {
		return []Playlist{}, ErrNoAudioPlaylists
	}

	return audioPlaylists, err
}

// GetPlaylistItems fetches detailed information about the playlist contents.
func (c Client) GetPlaylistItems(ctx context.Context, playlistID string) ([]PlaylistItem, error) {
	resp, err := api.FetchJSON[getPlaylistItemsResponse](
		ctx,
		c.httpClient,
		c.url(fmt.Sprintf("playlists/%s/items", playlistID)),
		c.log,
	)
	if err != nil {
		return []PlaylistItem{}, err
	}

	if resp.MediaContainer.Size == 0 {
		return []PlaylistItem{}, ErrNoItemsInPlaylist
	}

	return resp.MediaContainer.Metadata, nil
}

func (c Client) url(path string) string {
	return fmt.Sprintf("%s/%s?X-Plex-Token=%s", c.serverURL, path, c.authToken)
}

// getPlaylistsResponse is the response wrapper for the GetPlaylists request to align
// with how plex returns the data.
type getPlaylistsResponse struct {
	MediaContainer struct {
		Size     int        `json:"size,omitempty"`
		Metadata []Playlist `json:"metadata,omitempty"`
	} `json:"mediaContainer"`
}

// getPlaylistItemResponse is the response wrapper for the GetPlaylistItems request to align
// with how plex returns the data.
type getPlaylistItemsResponse struct {
	MediaContainer struct {
		Size     int            `json:"size,omitempty"`
		Metadata []PlaylistItem `json:"metadata,omitempty"`
	} `json:"mediaContainer"`
}
