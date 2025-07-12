package plex

// Error is a convenience type for setting error constants.
type Error uint8

const (
	// ErrNoPlaylists is the error returned when no playlists are found on the server.
	ErrNoPlaylists Error = iota + 1
	// ErrNoAudioPlaylists is the error returned when no audio playlists are found.
	ErrNoAudioPlaylists
	// ErrNoItemsInPlaylist is the error returned when no items are present in the playlist.
	ErrNoItemsInPlaylist
)

// Error returns the error message as string.
func (e Error) Error() string {
	switch e {
	case ErrNoPlaylists:
		return "no playlists found"

	case ErrNoAudioPlaylists:
		return "no audio playlists found"

	case ErrNoItemsInPlaylist:
		return "no items in playlist"

	default:
		return "unknown error"
	}
}
