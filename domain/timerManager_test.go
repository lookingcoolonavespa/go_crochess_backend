package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimerManager_StartTimer(t *testing.T) {
	timerManager := NewTimerManager()

	t.Run("Callback runs when timer ends", func(t *testing.T) {
		timerID := "1"
		duration, err := time.ParseDuration("1s")
		assert.NoError(t, err)

		successMsg := "callback fired"
		channel := make(chan string)
		timerManager.StartTimer(timerID, duration, func() {
			channel <- successMsg
		})

		select {
		case msg := <-channel:
			assert.Equal(t, msg, successMsg)
		}
	})
}

func TestTimerManager_DeleteTimer(t *testing.T) {
	timerManager := NewTimerManager()

	t.Run("Success", func(t *testing.T) {
		timerID := "1"
		duration, err := time.ParseDuration("1s")
		assert.NoError(t, err)

		successMsg := "callback fired"
		channel := make(chan string)
		timerManager.StartTimer(timerID, duration, func() {
			channel <- successMsg
		})

		err = timerManager.StopAndDeleteTimer(timerID)
		assert.NoError(t, err)

		failedMsg := "callback never fired"

		time.AfterFunc(duration*2, func() { channel <- failedMsg })

		select {
		case msg := <-channel:
			assert.Equal(t, msg, failedMsg)
		}
	})

	t.Run("Failed", func(t *testing.T) {
		timerID := "1"
		duration, err := time.ParseDuration("1s")
		assert.NoError(t, err)

		timerManager.StartTimer(timerID, duration, func() {})

		err = timerManager.StopAndDeleteTimer("2")
		assert.Error(t, err)
	})
}
