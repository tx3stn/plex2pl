package plex

// Error is a convenience type for setting error constants.
type Error uint8

const (
	// ErrNoPlaylists is the error returned when no playlists are found on the server.
	ErrNoPlaylists Error = iota + 1
)

// Error returns the error message as string.
func (e Error) Error() string {
	switch e {
	case ErrNoPlaylists:
		return "no playlists found"

	default:
		return "unknown error"
	}
}
