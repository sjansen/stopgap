package testutil_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sjansen/stopgap/internal/testutil"
	"github.com/sjansen/stopgap/internal/time"
)

func TestClock(t *testing.T) {
	require := require.New(t)

	c := &testutil.Clock{}
	require.Equal(0*time.Second, c.Paused)

	c.Sleep(5 * time.Second)
	require.Equal(5*time.Second, c.Paused)

	c.Sleep(5 * time.Second)
	require.Equal(10*time.Second, c.Paused)
}
