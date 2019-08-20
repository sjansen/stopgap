package mutex

import (
	"github.com/sjansen/stopgap/internal/rqx"
	"github.com/sjansen/stopgap/internal/time"
)

type Clock interface {
	Sleep(time.Duration)
}

type Repo interface {
	CreateMutex(rqx *rqx.RequestContext, name, description string) error
	LockMutex(rqx *rqx.RequestContext, name, message string) error
	UnlockMutex(rqx *rqx.RequestContext, name string) error
}

type Manager struct {
	Clock   Clock
	Mutexes Repo
}

var lockRetryDelays = []time.Duration{
	0 * time.Second,
	1 * time.Second,
	3 * time.Second,
	6 * time.Second,
	10 * time.Second,
}

func (m *Manager) CreateMutex(rqx *rqx.RequestContext, name, description string) error {
	return m.Mutexes.CreateMutex(rqx, name, description)
}

func (m *Manager) LockMutex(rqx *rqx.RequestContext, name, message string) error {
	var err error
	for _, d := range lockRetryDelays {
		m.Clock.Sleep(d)
		if err = m.Mutexes.LockMutex(rqx, name, message); err == nil {
			break
		}
	}
	return err
}
