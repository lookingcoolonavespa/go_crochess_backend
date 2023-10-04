package domain_websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockMessage struct {
	Topic   string `json:"topic"`
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

type testPayload struct {
	data string
}

const (
	testTopic = "topic"
)

func setupWebSocketRouter(t *testing.T, expected string) (WebSocketRouter, chan string, error) {
	r, err := NewWebSocketRouter()
	assert.NoError(t, err)

	topic, err := NewTopic(testTopic)
	assert.NoError(t, err)

	var message InboundMessage
	err = json.Unmarshal([]byte(expected), &message)
	assert.NoError(t, err)
	expectedPayload := message.Payload

	successChan := make(chan string)
	successStr := "success"
	mockHandleFunc := func(ctx context.Context, room *Room, client *Client, payload []byte) error {
		if string(payload) != string(expectedPayload) {
			return errors.New(fmt.Sprintf("expected payload: %v\nreceived payload: %v", expectedPayload[1], (payload)))
		}
		go func() {
			successChan <- successStr
		}()
		return nil
	}

	topic.RegisterEvent(SubscribeEvent, mockHandleFunc)

	r.PushNewRoute(topic)

	return r, successChan, nil
}

func TestWebSocketRouter_HandleWSMessage(t *testing.T) {
	tests := []struct {
		name      string
		expected  string
		shouldErr bool
	}{
		{
			name: "simple_pass",
			expected: fmt.Sprintf(`{"topic": "%s",
            "event": "%s",
            "payload": "test ran successfully"}`,
				testTopic,
				SubscribeEvent,
			),
			shouldErr: false,
		},
		{
			name: "invalid_json-event_field",
			expected: fmt.Sprintf(`{"topic": "%s",
            "payload": "test ran successfully"}`,
				testTopic,
			),
			shouldErr: true,
		},
		{
			name: "invalid_json-topic_field",
			expected: fmt.Sprintf(`{
            "event": "%s",
            "payload": "test ran successfully"}`,
				SubscribeEvent,
			),
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				r, successChan, err := setupWebSocketRouter(t, tt.expected)

				if err != nil {
					t.Errorf("error setting up WebSocketRouter: %v", err)
				}

				errChan := make(chan []byte)

				r.HandleWSMessage(
					context.Background(),
					NewClient(0, errChan, nil, nil),
					[]byte(tt.expected),
				)

				select {
				case err := <-errChan:
					if !tt.shouldErr {
						t.Errorf("error running HandleWSMessage: %v", string(err))
					}
				case <-successChan:
					if tt.shouldErr {
						t.Errorf("HandleWSMessage should err")
					}
				}
			},
		)
	}
}
