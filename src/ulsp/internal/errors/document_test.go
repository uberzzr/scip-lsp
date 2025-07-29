package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "document not found",
			err:  &DocumentNotFoundError{},
		},
		{
			name: "document size limit",
			err:  &DocumentSizeLimitError{},
		},
		{
			name: "document outdated",
			err:  &DocumentOutdatedError{},
		},
		{
			name: "invalid type",
			err:  &DocumentLanguageIDError{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.True(t, len(tt.err.Error()) > 0)
		})
	}
}
