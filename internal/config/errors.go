package config

// Error is a convenience type for creating error constants.
type Error uint8

const (
	// ErrConfigNotFound is the error returned when the config file cannot be found.
	ErrConfigNotFound Error = iota + 1
	// ErrReadingConfigFile is the error returned when reading the config file fails.
	ErrReadingConfigFile
	// ErrUnmarshalingConfig is the error returned when unmarshalling the config file fails.
	ErrUnmarshalingConfig
	// ErrInvalidOutputFormat is the error returned when the configured output format is
	// not a supported value.
	ErrInvalidOutputFormat
	// ErrMissingOutputFormat is the error returned when no output format is set in the
	// config file.
	ErrMissingOutputFormat
	// ErrMissingJellyfinServerURL is the error returned when the jellyfin output format
	// is set but no jellyfin server url is configured.
	ErrMissingJellyfinServerURL
	// ErrMissingJellyfinAPIKey is the error returned when the jellyfin output format is
	// set but no jellyfin API key is configured.
	ErrMissingJellyfinAPIKey
	// ErrMissingJellyfinOwnerUserID is the error returned when the jellyfin output format
	// is set but no jellyfin owner user id is configured.
	ErrMissingJellyfinOwnerUserID
)

// Error returns the error message string.
func (e Error) Error() string {
	switch e {
	case ErrConfigNotFound:
		return "config file not found"

	case ErrReadingConfigFile:
		return "error reading config file"

	case ErrUnmarshalingConfig:
		return "error unmarshaling config file"

	case ErrInvalidOutputFormat:
		return "invalid output format"

	case ErrMissingOutputFormat:
		return "no output format set in config file"

	case ErrMissingJellyfinServerURL:
		return "jellyfin output format requires jellyfinServerUrl to be set"

	case ErrMissingJellyfinAPIKey:
		return "jellyfin output format requires jellyfinApiKey to be set"

	case ErrMissingJellyfinOwnerUserID:
		return "jellyfin output format requires jellyfinOwnerUserId to be set"

	default:
		return "unknown error"
	}
}
