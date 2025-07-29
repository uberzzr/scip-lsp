package errors

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUUIDNotFound(t *testing.T) {
	id := uuid.Must(uuid.FromString("4d8c6b36-4e9b-4469-8a05-2c60b9671590"))
	err := &UUIDNotFoundError{UUID: id}
	msg := `UUID "4d8c6b36-4e9b-4469-8a05-2c60b9671590" not found`
	assert.Equal(t, msg, err.Error())
}

func TestIsUUIDNotFound(t *testing.T) {
	id := uuid.Must(uuid.NewV4())
	tests := []struct {
		name     string
		err      error
		wantOK   bool
		wantUUID uuid.UUID
	}{
		{
			name:     "uuid not found",
			err:      &UUIDNotFoundError{UUID: id},
			wantOK:   true,
			wantUUID: id,
		},
		{
			name:     "random error",
			err:      New("err"),
			wantOK:   false,
			wantUUID: uuid.Nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			id, ok := NotFoundUUID(tt.err)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantUUID, id)
		})
	}
}
