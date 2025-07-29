package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestUlspDaemonString(t *testing.T) {
	t.Run("nil ulspdaemon", func(t *testing.T) {
		var f *UlspDaemon
		assert.Empty(t, f.String(), "Unexpected string representation.")
	})
	t.Run("populated ulspdaemon", func(t *testing.T) {
		f := &UlspDaemon{Name: "mensch"}
		assert.Equal(
			t,
			"",
			f.String(),
			"Unexpected string representation.",
		)
	})
}

func TestUlspDaemonKeys(t *testing.T) {
	t.Run("nil ulspdaemon", func(t *testing.T) {
		var f *UlspDaemon
		assert.Equal(t, "request_ulspdaemon", f.RequestKey(), "Unexpected request key")
		assert.Equal(t, "response_ulspdaemon", f.ResponseKey(), "Unexpected response key")
	})
	t.Run("populated ulspdaemon", func(t *testing.T) {
		f := &UlspDaemon{}
		assert.Equal(t, "request_ulspdaemon", f.RequestKey(), "Unexpected request key")
		assert.Equal(t, "response_ulspdaemon", f.ResponseKey(), "Unexpected response key")
	})
}

func TestClientNameIsVSCodeBased(t *testing.T) {
	testCases := []struct {
		name     string
		client   ClientName
		expected bool
	}{
		{
			name:     "VS Code client",
			client:   ClientNameVSCode,
			expected: true,
		},
		{
			name:     "Cursor client",
			client:   ClientNameCursor,
			expected: true,
		},
		{
			name:     "Other client",
			client:   "Some Other Client",
			expected: false,
		},
		{
			name:     "Empty client name",
			client:   "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.client.IsVSCodeBased()
			assert.Equal(t, tc.expected, result, "Unexpected result for client %q", tc.client)
		})
	}
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
