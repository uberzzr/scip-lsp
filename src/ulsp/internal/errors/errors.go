package errors

import stderr "errors"

// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
func New(msg string) error {
	return stderr.New(msg)
}

var (
	// NoUUIDOnWireError reports that the request is missing a UUID.
	NoUUIDOnWireError = New("UUID is required")
	// NoMessageOnWireError reports that the request is missing a message.
	NoMessageOnWireError = New("no message on wire")
)

// IsBadRequest reports whether the error is a bad request from the caller.
func IsBadRequest(e error) bool {
	return stderr.Is(e, NoUUIDOnWireError) || stderr.Is(e, NoMessageOnWireError)
}
