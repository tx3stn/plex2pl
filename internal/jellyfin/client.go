package jellyfin

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tx3stn/plex2pl/internal/api"
	"github.com/tx3stn/plex2pl/internal/logger"
)

// mediaTypeAudio is the jellyfin media type value for audio playlists.
const mediaTypeAudio = "Audio"

// Client is the jellyfin client for making requests to your jellyfin server API.
type Client struct {
	httpClient api.HTTPClient
	serverURL  string
	apiKey     string
	log        *logger.Basic
}

// NewClient creates a new instance of the jellyfin API Client.
func NewClient(
	httpClient api.HTTPClient,
	serverURL string,
	apiKey string,
	log *logger.Basic,
) *Client {
	return &Client{
		httpClient: httpClient,
		serverURL:  serverURL,
		apiKey:     apiKey,
		log:        log,
	}
}

// AudioItemIndex fetches every audio item in the library and returns a map of file
// path to jellyfin item id, used to resolve plex track file paths to the jellyfin
// item ids required when creating a playlist.
func (c Client) AudioItemIndex(ctx context.Context) (map[string]string, error) {
	query := url.Values{}
	query.Set("Recursive", "true")
	query.Set("IncludeItemTypes", mediaTypeAudio)
	query.Set("Fields", "Path")
	query.Set("EnableImages", "false")

	resp, err := api.SendJSON[itemsResponse](
		ctx,
		c.httpClient,
		http.MethodGet,
		c.url("Items", query),
		nil,
		c.headers(),
		c.log,
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching jellyfin audio items: %w", err)
	}

	index := make(map[string]string, len(resp.Items))

	for _, item := range resp.Items {
		if item.Path == "" || item.ID == "" {
			continue
		}

		index[item.Path] = item.ID
	}

	return index, nil
}

// PlaylistNames fetches the names of the playlists that already exist for the owner,
// used to skip re-creating playlists that are already present in jellyfin.
func (c Client) PlaylistNames(
	ctx context.Context,
	ownerUserID string,
) (map[string]struct{}, error) {
	query := url.Values{}
	query.Set("Recursive", "true")
	query.Set("IncludeItemTypes", "Playlist")

	if ownerUserID != "" {
		query.Set("UserId", ownerUserID)
	}

	resp, err := api.SendJSON[itemsResponse](
		ctx,
		c.httpClient,
		http.MethodGet,
		c.url("Items", query),
		nil,
		c.headers(),
		c.log,
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching jellyfin playlists: %w", err)
	}

	names := make(map[string]struct{}, len(resp.Items))

	for _, item := range resp.Items {
		if item.Name == "" {
			continue
		}

		names[item.Name] = struct{}{}
	}

	return names, nil
}

// CreatePlaylist creates a new jellyfin playlist owned by ownerUserID containing the
// given jellyfin item ids, returning the id of the created playlist.
func (c Client) CreatePlaylist(
	ctx context.Context,
	name string,
	itemIDs []string,
	ownerUserID string,
) (string, error) {
	body := createPlaylistRequest{
		Name:      name,
		IDs:       itemIDs,
		UserID:    ownerUserID,
		MediaType: mediaTypeAudio,
	}

	resp, err := api.SendJSON[createPlaylistResponse](
		ctx,
		c.httpClient,
		http.MethodPost,
		c.url("Playlists", nil),
		body,
		c.headers(),
		c.log,
	)
	if err != nil {
		return "", fmt.Errorf("error creating jellyfin playlist '%s': %w", name, err)
	}

	return resp.ID, nil
}

// headers returns the request headers required to authenticate against the jellyfin API.
func (c Client) headers() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("MediaBrowser Token=%q", c.apiKey),
	}
}

// url builds a request url for the given API path and optional query parameters.
func (c Client) url(path string, query url.Values) string {
	requestURL := fmt.Sprintf("%s/%s", strings.TrimRight(c.serverURL, "/"), path)

	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}

	return requestURL
}

// itemsResponse mirrors the jellyfin /Items response. Jellyfin uses PascalCase keys.
type itemsResponse struct {
	Items []item `json:"Items"`
}

// item is a single entry in the jellyfin /Items response.
type item struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
	Path string `json:"Path"`
}

// createPlaylistRequest is the CreatePlaylistDto payload for POST /Playlists.
type createPlaylistRequest struct {
	Name      string   `json:"Name"`
	IDs       []string `json:"Ids"`
	UserID    string   `json:"UserId"`
	MediaType string   `json:"MediaType"`
}

// createPlaylistResponse is the PlaylistCreationResult returned by POST /Playlists.
type createPlaylistResponse struct {
	ID string `json:"Id"`
}
