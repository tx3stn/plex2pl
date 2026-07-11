package api

// Error is a convenience type for the error constants.
type Error uint8

const (
	// ErrMashallingJSONToType is the error returned when fetch can't marshall the response JSON
	// from a request to the provided data type.
	ErrMashallingJSONToType Error = iota + 1
	// ErrMakingRequest is the error returned when the http request can't be made.
	ErrMakingRequest
	// ErrUnexpectedStatus is the error returned when a request responds with a non-2xx
	// status code.
	ErrUnexpectedStatus
)

func (e Error) Error() string {
	switch e {
	case ErrMashallingJSONToType:
		return "error marshalling response json to type"
	case ErrMakingRequest:
		return "error making request"
	case ErrUnexpectedStatus:
		return "unexpected response status code"
	default:
		return "unknown error"
	}
}
