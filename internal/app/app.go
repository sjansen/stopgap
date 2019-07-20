package app

import (
	"github.com/sjansen/stopgap/internal/rqx"
	"github.com/sjansen/stopgap/internal/time"
)

type Clock interface {
	Sleep(time.Duration)
}

type MutexRepo interface {
	Create(rqx *rqx.RequestContext, name, description string) error
	Lock(rqx *rqx.RequestContext, name, message string) error
	Unlock(rqx *rqx.RequestContext, name, message string) error
}

type App struct {
	Clock   Clock
	Mutexes MutexRepo
}

var lockRetryDelays = []time.Duration{
	0 * time.Second,
	1 * time.Second,
	3 * time.Second,
	6 * time.Second,
	10 * time.Second,
}

func (a *App) CreateMutex(rqx *rqx.RequestContext, name, description string) error {
	return a.Mutexes.Create(rqx, name, description)
}

func (a *App) LockMutex(rqx *rqx.RequestContext, name, message string) error {
	var err error
	for _, d := range lockRetryDelays {
		a.Clock.Sleep(d)
		if err = a.Mutexes.Lock(rqx, name, message); err == nil {
			break
		}
	}
	return err
}
