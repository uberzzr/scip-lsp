package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMeasure(t *testing.T) {
	model := UlspDaemon{Name: "TestName"}
	assert.Equal(t, model.Name, "TestName")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
