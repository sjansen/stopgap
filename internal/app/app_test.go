package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sjansen/stopgap/internal/app"
	"github.com/sjansen/stopgap/internal/rqx"
	"github.com/sjansen/stopgap/internal/storage"
	"github.com/sjansen/stopgap/internal/testutil"
	"github.com/sjansen/stopgap/internal/time"
)

func TestCreateMutex(t *testing.T) {
	require := require.New(t)

	tc := newTestCase()
	// GIVEN an unused mutex name
	require.NotContains(tc.repo.Mutexes, "triton")
	// WHEN there is an attempt to create the mutex
	err := tc.app.CreateMutex(tc.rqx, "triton", "staging and prod")
	// THEN mutex and its description should be added to the repo
	require.Equal(tc.repo.Mutexes["triton"], "staging and prod")
	// and there shouldn't be an error
	require.NoError(err)
}

func TestLockMutex(t *testing.T) {
	require := require.New(t)

	tc := newTestCase()
	// GIVEN a mutex that will be unlocked soon
	tc.repo.Retries = 5
	// WHEN there is an attempt to lock the mutex
	err := tc.app.LockMutex(tc.rqx, "triton", "rebooting the world")
	// THEN retrying should have taken 20 seconds
	require.Equal(20*time.Second, tc.clock.Paused)
	// and there shouldn't be an error
	require.NoError(err)
}

type dependencies struct {
	app   *app.App
	rqx   *rqx.RequestContext
	clock *testutil.Clock
	repo  *storage.MutexRepoFake
}

func newTestCase() *dependencies {
	clock := &testutil.Clock{}
	repo := storage.NewMutexRepoFake()
	return &dependencies{
		clock: clock,
		repo:  repo,
		app: &app.App{
			Clock:   clock,
			Mutexes: repo,
		},
		rqx: &rqx.RequestContext{
			Ctx: context.TODO(),
		},
	}
}
