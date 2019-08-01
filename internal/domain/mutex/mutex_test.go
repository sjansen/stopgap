package mutex_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sjansen/stopgap/internal/domain/mutex"
	"github.com/sjansen/stopgap/internal/rqx"
	"github.com/sjansen/stopgap/internal/storage"
	"github.com/sjansen/stopgap/internal/testutil"
	"github.com/sjansen/stopgap/internal/time"
)

func TestCreateMutex(t *testing.T) {
	require := require.New(t)

	deps := newDependencies()
	// GIVEN an unused mutex name
	require.NotContains(deps.repo.Mutexes, "triton")
	// WHEN there is an attempt to create the mutex
	err := deps.manager.CreateMutex(deps.rqx, "triton", "staging and prod")
	// THEN the mutex should be added to the repo
	require.Equal(deps.repo.Mutexes["triton"], "staging and prod")
	require.NoError(err)
}

func TestLockMutex(t *testing.T) {
	require := require.New(t)

	deps := newDependencies()
	// GIVEN a mutex that will be unlocked soon
	deps.repo.Retries = 5
	// WHEN there is an attempt to lock the mutex
	err := deps.manager.LockMutex(deps.rqx, "triton", "rebooting the world")
	// THEN it should succeed after retrying for 20 seconds
	require.Equal(20*time.Second, deps.clock.Paused)
	require.NoError(err)
}

type dependencies struct {
	manager *mutex.Manager
	rqx     *rqx.RequestContext
	clock   *testutil.Clock
	repo    *storage.MutexRepoFake
}

func newDependencies() *dependencies {
	clock := &testutil.Clock{}
	repo := storage.NewMutexRepoFake()
	return &dependencies{
		clock: clock,
		repo:  repo,
		manager: &mutex.Manager{
			Clock:   clock,
			Mutexes: repo,
		},
		rqx: &rqx.RequestContext{
			Ctx: context.TODO(),
		},
	}
}
