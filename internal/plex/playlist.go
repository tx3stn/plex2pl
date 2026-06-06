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

// PlaylistItem is the strcuture of the detailed information about a playlist.
type PlaylistItem struct {
	RatingKey        string          `json:"ratingKey,omitempty"`
	Title            string          `json:"title,omitempty"`
	GrandParentTitle string          `json:"grandParentTitle,omitempty"`
	ParentTitle      string          `json:"parentTitle,omitempty"`
	Genre            []Tag           `json:"genre,omitempty"`
	Media            []PlaylistMedia `json:"media,omitempty"`
}

// Tag represents the tag entries plex includes in metadata responses, e.g. the
// 'Genre' section of a track's metadata.
type Tag struct {
	Tag string `json:"tag,omitempty"`
}

// PlaylistMedia represents the data in the 'Media' section of the 'Metadata' for the
// playlist item.
type PlaylistMedia struct {
	ID               int         `json:"id,omitempty"`
	Duration         int         `json:"duration,omitempty"`
	Bitrate          int         `json:"bitrate,omitempty"`
	AudioChannels    int         `json:"audioChannels,omitempty"`
	AudioCodec       string      `json:"audioCodec,omitempty"`
	Container        string      `json:"container,omitempty"`
	HasVoiceActivity bool        `json:"hasVoiceActivity,omitempty"`
	Part             []MediaPart `json:"part,omitempty"`
}

// MediaPart represents the data in the 'Part' setion of the 'Media' file in the
// playlist item.
type MediaPart struct {
	ID           int    `json:"id,omitempty"`
	Key          string `json:"key,omitempty"`
	Duration     int    `json:"duration,omitempty"`
	File         string `json:"file,omitempty"`
	Size         int    `json:"size,omitempty"`
	Container    string `json:"container,omitempty"`
	HasThumbnail string `json:"hasThumbnail,omitempty"`
}
