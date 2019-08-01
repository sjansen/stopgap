package mutex

import (
	"github.com/sjansen/stopgap/internal/rqx"
	"github.com/sjansen/stopgap/internal/time"
)

type Clock interface {
	Sleep(time.Duration)
}

type Repo interface {
	Create(rqx *rqx.RequestContext, name, description string) error
	Lock(rqx *rqx.RequestContext, name, message string) error
	Unlock(rqx *rqx.RequestContext, name, message string) error
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
	return m.Mutexes.Create(rqx, name, description)
}

func (m *Manager) LockMutex(rqx *rqx.RequestContext, name, message string) error {
	var err error
	for _, d := range lockRetryDelays {
		m.Clock.Sleep(d)
		if err = m.Mutexes.Lock(rqx, name, message); err == nil {
			break
		}
	}
	return err
}
