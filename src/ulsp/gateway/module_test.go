package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/goleak"
)

func TestModule(t *testing.T) {
	assert.NotPanics(t, func() { fx.Provide() })
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
