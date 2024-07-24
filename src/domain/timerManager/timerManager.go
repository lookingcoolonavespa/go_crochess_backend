package domain_timerManager

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type TimerManager struct {
	timers map[int]*time.Timer
	mutex  sync.Mutex
}

func NewTimerManager() *TimerManager {
	return &TimerManager{
		make(map[int]*time.Timer),
		sync.Mutex{},
	}
}

func (t *TimerManager) StartTimer(timerID int, duration time.Duration, callback func()) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	existingTimer, exists := t.timers[timerID]
	if exists {
		existingTimer.Stop()
	}

	newTimer := time.AfterFunc(duration, func() {
		callback()
		t.StopAndDeleteTimer(timerID)
	},
	)

	t.timers[timerID] = newTimer
}

func (t *TimerManager) StopAndDeleteTimer(timerID int) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	timer, exists := t.timers[timerID]
	if exists {
		timer.Stop()
		delete(t.timers, timerID)
		return nil
	} else {
		return errors.New(fmt.Sprintf("There was no timer at %d.", timerID))
	}
}
