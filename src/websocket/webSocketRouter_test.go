package domain_websocket

import (
	"context"
	"errors"
	"fmt"
	"testing"

	domain_websocket_mock "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket/mock"
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

	payloadRegex, err := jsonRegex("payload")
	assert.NoError(t, err)

	expectedPayload := payloadRegex.FindStringSubmatch(expected)
	if len(expectedPayload) != 2 {
		return WebSocketRouter{}, nil, errors.New("payload field is missing in expected")
	}

	successChan := make(chan string)
	successStr := "success"
	mockHandleFunc := func(ctx context.Context, room *Room, client Client, payload []byte) error {
		if string(payload) != expectedPayload[1] {
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
					domain_websocket_mock.NewMockClient(errChan),
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
