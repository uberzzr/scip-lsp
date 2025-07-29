package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.NotNil(t, New())
}

func TestSleep(t *testing.T) {
	assert.NotPanics(t, func() {
		clock{}.Sleep(1 * time.Microsecond)
	})
}
