package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBadRequest(t *testing.T) {
	nb := New("not bad request")
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "no uuid on wire",
			err:  NoUUIDOnWireError,
			want: true,
		},
		{
			name: "no message on wire",
			err:  NoMessageOnWireError,
			want: true,
		},
		{
			name: "not bad request",
			err:  nb,
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsBadRequest(tt.err))
		})
	}
}
