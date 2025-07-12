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

	default:
		return "unknown error"
	}
}
