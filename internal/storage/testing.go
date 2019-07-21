package storage

import (
	"errors"

	"github.com/sjansen/stopgap/internal/rqx"
)

type MutexRepoFake struct {
	Retries int
	Mutexes map[string]string
}

func NewMutexRepoFake() *MutexRepoFake {
	return &MutexRepoFake{
		Retries: 0,
		Mutexes: map[string]string{"conch": "migrations"},
	}
}

func (r *MutexRepoFake) Create(rqx *rqx.RequestContext, name, description string) error {
	r.Mutexes[name] = description
	return nil
}

func (r *MutexRepoFake) Lock(rqx *rqx.RequestContext, name, message string) error {
	r.Retries--
	if r.Retries > 0 {
		return errors.New("already locked")
	}
	return nil
}

func (r *MutexRepoFake) Unlock(rqx *rqx.RequestContext, name, message string) error {
	return nil
}
