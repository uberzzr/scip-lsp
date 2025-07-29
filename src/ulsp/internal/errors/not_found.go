package errors

import (
	stderr "errors"
	"fmt"

	"github.com/gofrs/uuid"
)

// UUIDNotFoundError is a service domain error for not found.
type UUIDNotFoundError struct {
	UUID uuid.UUID
}

// Error is an implementation of the error interface.
func (n *UUIDNotFoundError) Error() string {
	return fmt.Sprintf("UUID %q not found", n.UUID)
}

// NotFoundUUID returns an UUID and true if UUIDNotFoundError is part of the
// error chain.
func NotFoundUUID(e error) (_ uuid.UUID, ok bool) {
	var nf *UUIDNotFoundError
	if !stderr.As(e, &nf) {
		return uuid.Nil, false
	}
	return nf.UUID, true
}

// NoSessionFoundError indicates that a session cannot be found within the context.
type NoSessionFoundError struct{}

// Error is an implementation of the error interface.
func (n *NoSessionFoundError) Error() string {
	return fmt.Sprintf("No session found in context")
}
