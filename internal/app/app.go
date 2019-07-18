package app

import (
	"github.com/sjansen/stopgap/internal/time"
)

type Clock interface {
	Sleep(time.Duration)
}

type MutexRepo interface {
	Create(name, description string) error
	Lock(name, message string) error
	Unlock(name, message string) error
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
	10 * time.Second,
}

func (a *App) CreateMutex(name, description string) error {
	return a.Mutexes.Create(name, description)
}

func (a *App) LockMutex(name, message string) error {
	var err error
	for _, d := range lockRetryDelays {
		a.Clock.Sleep(d)
		if err = a.Mutexes.Lock(name, message); err == nil {
			break
		}
	}
	return err
}
