package domain_timerManager

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type TimerManager struct {
	timers  sync.Map
	mutexes map[string]*sync.Mutex
}

func NewTimerManager() TimerManager {
	return TimerManager{
		timers:  sync.Map{},
		mutexes: make(map[string]*sync.Mutex),
	}
}

func (t *TimerManager) StartTimer(timerID string, duration time.Duration, callback func()) {
	newTimer := time.AfterFunc(duration, func() {
		if timer, exists := t.timers.LoadAndDelete(timerID); exists {
			callback()
			newTimer := timer.(*time.Timer)
			newTimer.Stop()
		}
	},
	)

	existingTimer, exists := t.timers.Swap(timerID, newTimer)
	if exists {
		timer := existingTimer.(*time.Timer)
		timer.Stop()
	}
}

func (t *TimerManager) StopAndDeleteTimer(timerID string) error {
	timer, exists := t.timers.LoadAndDelete(timerID)
	if exists {
		timer.(*time.Timer).Stop()
		return nil
	} else {
		return errors.New(fmt.Sprintf("There was no timer at %s.", timerID))
	}
}
