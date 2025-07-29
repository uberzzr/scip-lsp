package clock

import (
	"time"
)

// Clock is an interface that abstracts the functionality for measuring and displaying time.
type Clock interface {
	// Sleep pauses the current goroutine for at least the duration d. A negative or zero duration causes Sleep to return immediately.
	Sleep(duration time.Duration)
}

type clock struct{}

// New creates a new instance of Clock.
func New() Clock {
	return clock{}
}

func (clock) Sleep(duration time.Duration) {
	time.Sleep(duration)
}
