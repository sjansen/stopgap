package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sjansen/stopgap/internal/app"
	"github.com/sjansen/stopgap/internal/rqx"
	"github.com/sjansen/stopgap/internal/testutil"
	"github.com/sjansen/stopgap/internal/time"
)

func TestCreateMutex(t *testing.T) {
	require := require.New(t)

	// GIVEN an unused mutex name
	repo := &mutexRepo{
		mutexes: map[string]string{"conch": "migrations"},
	}
	// and any other dependencies
	app := &app.App{
		Clock:   &testutil.Clock{},
		Mutexes: repo,
	}
	rqx := &rqx.RequestContext{
		Ctx: context.TODO(),
	}

	// WHEN there is an attempt to create the mutex
	err := app.CreateMutex(rqx, "triton", "staging and prod")

	// THEN there shouldn't be an error
	require.NoError(err)
	// and the mutex and its description should be added to the repo
	require.Equal(repo.mutexes["triton"], "staging and prod")
}

func TestLockMutex(t *testing.T) {
	require := require.New(t)

	// GIVEN a mutex that is already locked
	repo := &mutexRepo{retries: 5}
	// and any other dependencies
	clock := &testutil.Clock{}
	app := &app.App{
		Clock:   clock,
		Mutexes: repo,
	}
	rqx := &rqx.RequestContext{
		Ctx: context.TODO(),
	}

	// WHEN there is an attempt to lock the mutex
	err := app.LockMutex(rqx, "triton", "rebooting the world")
	// and retrying succeeds within 20 seconds
	require.Equal(20*time.Second, clock.Paused)

	// THEN there shouldn't be an error
	require.NoError(err)
}

type mutexRepo struct {
	retries int
	mutexes map[string]string
}

func (r *mutexRepo) Create(rqx *rqx.RequestContext, name, description string) error {
	r.mutexes[name] = description
	return nil
}

func (r *mutexRepo) Lock(rqx *rqx.RequestContext, name, message string) error {
	r.retries--
	if r.retries > 0 {
		return errors.New("already locked")
	}
	return nil
}

func (r *mutexRepo) Unlock(rqx *rqx.RequestContext, name, message string) error {
	return nil
}
