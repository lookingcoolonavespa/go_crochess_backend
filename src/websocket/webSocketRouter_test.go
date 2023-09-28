package domain_websocket

import (
	"context"
	"errors"
	"fmt"
	"testing"

	domain_websocket_mock "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket/mock"
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
	testEvent = "run test"
	testTopic = "topic"
)

func setupWebSocketRouter(t *testing.T, expected string) (*WebSocketRouter, error) {
	r, err := NewWebSocketRouter()
	if err != nil {
		return nil, err
	}

	topic, err := NewTopic(testTopic)
	if err != nil {
		return nil, err
	}

	payloadRegex, err := jsonRegex("payload")
	if err != nil {
		return nil, err
	}
	expectedPayload := payloadRegex.FindStringSubmatch(expected)
	if len(expectedPayload) != 2 {
		return nil, errors.New("payload field is missing in expected")
	}

	mockHandleFunc := func(ctx context.Context, room Room, client Client, payload []byte) error {
		if string(payload) != expectedPayload[1] {
			return errors.New(fmt.Sprintf("expected payload: %v\nreceived payload: %v", expectedPayload[1], (payload)))
		}
		return nil
	}

	topic.RegisterEvent(testEvent, mockHandleFunc)

	r.PushNewRoute(topic)

	return r, nil
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
				testEvent,
			),
			shouldErr: false,
		},
		{
			name: "invalid_json-event_field",
			expected: fmt.Sprintf(`{"topic": "%s",
            "event" "%s",
            "payload": "test ran successfully"}`,
				testTopic,
				testEvent,
			),
			shouldErr: true,
		},
		{
			name: "invalid_json-topic_field",
			expected: fmt.Sprintf(`{
            "event": "%s",
            "payload": "test ran successfully"}`,
				testEvent,
			),
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				r, err := setupWebSocketRouter(t, tt.expected)

				if err != nil {
					t.Errorf("error setting up WebSocketRouter: %v", err)
				}

				err = r.HandleWSMessage(
					context.Background(),
					domain_websocket_mock.NewMockClient(make(chan []byte)),
					[]byte(tt.expected),
					len(tt.expected),
				)
				if err != nil && !tt.shouldErr {
					t.Errorf("HandleWSMessage should err")
				} else if err == nil && tt.shouldErr {
					t.Errorf("error running HandleWSMessage: %v", err)
				}
			},
		)
	}
}
