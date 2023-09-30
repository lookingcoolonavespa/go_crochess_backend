package domain_websocket

import (
	"context"
	"testing"

	domain_websocket_mock "github.com/lookingcoolonavespa/go_crochess_backend/src/websocket/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewTopic(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		shouldError bool
	}{
		{
			name:        "only topic - success",
			pattern:     "topic",
			shouldError: false,
		},
		{
			name:        "only topic - failure because of whitespace",
			pattern:     "topic  ",
			shouldError: true,
		},
		{
			name:        "only topic - failure because of special character",
			pattern:     "topic*",
			shouldError: true,
		},
		{
			name:        "topic and pattern - success",
			pattern:     "topic/param",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				_, err := NewTopic(tt.pattern)
				if err == nil && tt.shouldError {
					t.Errorf("pattern should cause an error\npattern: %s", tt.pattern)
				} else if err != nil && !tt.shouldError {
					t.Errorf("pattern shouldn't cause an error\npattern: %s", tt.pattern)
				}
			},
		)
	}
}

func TestTopic_matcher(t *testing.T) {
	tests := []struct {
		name        string
		basePattern string
		testPattern []byte
		expected    bool
	}{
		{
			name:        "base:topic,test:topic",
			basePattern: "topic",
			testPattern: []byte("topic"),
			expected:    true,
		},
		{
			name:        "base:topic,test:topic/param",
			basePattern: "topic",
			testPattern: []byte("topic/param"),
			expected:    false,
		},
		{
			name:        "base:topic,test:topic/",
			basePattern: "topic",
			testPattern: []byte("topic/"),
			expected:    false,
		},
		{
			name:        "base:topic,test:topica",
			basePattern: "topic",
			testPattern: []byte("topica"),
			expected:    false,
		},
		{
			name:        "base:topic/param,test:topic/param",
			basePattern: "topic/param",
			testPattern: []byte("topic/param"),
			expected:    true,
		},
		{
			name:        "base:topic/param,test:topic/",
			basePattern: "topic/param",
			testPattern: []byte("topic/"),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				topic, err := NewTopic(tt.basePattern)
				if err != nil {
					t.Errorf("base pattern caused an error\nbase pattern: %s", tt.basePattern)
				}

				match := topic.match(tt.testPattern)
				if match != tt.expected {
					t.Fail()
				}
			},
		)
	}
}

func TestTopic_TopicWithoutParam_HandleWSMessage(t *testing.T) {
	type MessageParams struct {
		Pattern string
		Event   string
	}
	tests := []struct {
		name       string
		baseParams MessageParams
		testParams MessageParams
		shouldErr  bool
	}{
		{
			name:       "success",
			baseParams: MessageParams{"topic", SubscribeEvent},
			testParams: MessageParams{"topic", SubscribeEvent},
			shouldErr:  false,
		},
		{
			name:       "fail on not subscribed",
			baseParams: MessageParams{"topic", "event"},
			testParams: MessageParams{"topic", "event"},
			shouldErr:  true,
		},
		{
			name:       "fail on invalid event",
			baseParams: MessageParams{"topic", SubscribeEvent},
			testParams: MessageParams{"topic", "event"},
			shouldErr:  true,
		},
	}

	successStr := "success"
	setupTest := func(pattern string, event string) (Topic, chan string) {
		topic, err := NewTopic(pattern)
		assert.NoError(t, err)

		msgChan := make(chan string)
		handleFunc := func(context.Context, *Room, Client, []byte) error {
			go func() {
				msgChan <- successStr
			}()
			return nil
		}
		topic.RegisterEvent(event, handleFunc)

		return topic, msgChan
	}

	clientChan := make(chan []byte)
	client := domain_websocket_mock.NewMockClient(clientChan)
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				topic, msgChan := setupTest(tt.baseParams.Pattern, tt.baseParams.Event)
				topic.HandleWSMessage(
					context.Background(),
					client,
					tt.testParams.Event,
					[]byte{},
					[]byte(tt.testParams.Pattern),
				)

				select {
				case <-msgChan:
					assert.False(t, tt.shouldErr)
				case clientMsg := <-clientChan:
					assert.True(t, tt.shouldErr)
					assert.Contains(t, string(clientMsg), "error")
				}

			},
		)
	}
}

func TestTopic_TopicWithParam_HandleWSMessage(t *testing.T) {
	type MessageParams struct {
		Pattern string
		Event   string
	}
	tests := []struct {
		name       string
		baseParams MessageParams
		testParams MessageParams
		shouldErr  bool
	}{
		{
			name:       "success",
			baseParams: MessageParams{"topic/param", SubscribeEvent},
			testParams: MessageParams{"topic/param", SubscribeEvent},
			shouldErr:  false,
		},
		{
			name:       "fail on not subscribed",
			baseParams: MessageParams{"topic/param", SubscribeEvent},
			testParams: MessageParams{"topic/paramb", "event"},
			shouldErr:  true,
		},
		{
			name:       "fail on invalid event",
			baseParams: MessageParams{"topic/param", SubscribeEvent},
			testParams: MessageParams{"topic/param", "event"},
			shouldErr:  true,
		},
	}

	successStr := "success"
	setupTest := func(pattern string, event string) (Topic, chan string) {
		topic, err := NewTopic(pattern)
		assert.NoError(t, err)

		msgChan := make(chan string)
		handleFunc := func(context.Context, *Room, Client, []byte) error {
			go func() {
				msgChan <- successStr
			}()
			return nil
		}
		topic.RegisterEvent(event, handleFunc)

		return topic, msgChan
	}

	clientChan := make(chan []byte)
	client := domain_websocket_mock.NewMockClient(clientChan)
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				topic, msgChan := setupTest(tt.baseParams.Pattern, tt.baseParams.Event)
				topic.HandleWSMessage(
					context.Background(),
					client,
					tt.testParams.Event,
					[]byte{},
					[]byte(tt.testParams.Pattern),
				)

				select {
				case <-msgChan:
					assert.False(t, tt.shouldErr)
				case clientMsg := <-clientChan:
					assert.True(t, tt.shouldErr)
					assert.Contains(t, string(clientMsg), "error")
				}

			},
		)
	}
}
