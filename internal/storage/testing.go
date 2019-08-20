package storage

import (
	"errors"

	"github.com/sjansen/stopgap/internal/rqx"
)

type MutexRepoFake struct {
	Retries int
	Mutexes map[string]string
}

// NewMutexRepoFake creates a DynamoStore instance using default values.
func NewMutexRepoFake() *MutexRepoFake {
	return &MutexRepoFake{
		Retries: 0,
		Mutexes: map[string]string{"conch": "migrations"},
	}
}

// CreateMutex adds the named mutex.
func (r *MutexRepoFake) CreateMutex(rqx *rqx.RequestContext, name, description string) error {
	r.Mutexes[name] = description
	return nil
}

// LockMutex locks the named mutex.
func (r *MutexRepoFake) LockMutex(rqx *rqx.RequestContext, name, message string) error {
	r.Retries--
	if r.Retries > 0 {
		return errors.New("already locked")
	}
	return nil
}

// UnlockMutex unlocks the named mutex.
func (r *MutexRepoFake) UnlockMutex(rqx *rqx.RequestContext, name string) error {
	return nil
}
