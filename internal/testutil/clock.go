package testutil

import "github.com/sjansen/stopgap/internal/time"

type Clock struct {
	Paused time.Duration
}

func (c *Clock) Sleep(d time.Duration) {
	c.Paused += d
}
