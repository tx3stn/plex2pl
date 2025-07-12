package plex

// Playlist represents the playlist summary data.
type Playlist struct {
	RatingKey    string `json:"ratingKey,omitempty"`
	Key          string `json:"key,omitempty"`
	GUID         string `json:"guid,omitempty"`
	ItemType     string `json:"itemType,omitempty"`
	Title        string `json:"title,omitempty"`
	TitleSort    string `json:"titleSort,omitempty"`
	Summary      string `json:"summary,omitempty"`
	Smart        bool   `json:"smart,omitempty"`
	PlaylistType string `json:"playlistType,omitempty"`
	Icon         string `json:"icon,omitempty"`
	ViewCount    int    `json:"viewCount,omitempty"`
	LastViewedAt int    `json:"lastViewedAt,omitempty"`
	LeafCount    int    `json:"leafCount,omitempty"`
	AddedAt      int    `json:"addedAt,omitempty"`
	UpdatedAt    int    `json:"updatedAt,omitempty"`
}

// MediaContainer is the structure that wraps the list of playlists.
type MediaContainer struct {
	Size     int        `json:"size,omitempty"`
	Metadata []Playlist `json:"metadata,omitempty"`
}
