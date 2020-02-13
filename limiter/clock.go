package limiter

import (
	"sync"
	"time"
)

// A Clock is used to check if the day switched since the last call.
type Clock interface {
	DaySwitched() bool
}

type clock struct {
	lock     sync.Locker
	lastCall *time.Time
}

// NewClock Constructor.
func NewClock() Clock {
	return &clock{
		lock: &sync.Mutex{},
	}
}

func (cl *clock) DaySwitched() bool {
	today := roundToDay(time.Now())
	cl.lock.Lock()
	last := cl.lastCall
	cl.lastCall = &today
	result := last != nil && today.After(*last)
	cl.lock.Unlock()
	return result
}

func roundToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
