package domain_websocket

import "testing"

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
			expected:    true,
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

// func TestTopic_HandleWSMessage(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		basePattern string
// 		testPattern []byte
// 		expected    string
// 	}{
// 		{
// 			name:        "success",
// 			basePattern: "topic",
// 			testPattern: []byte("topic/param"),
// 			expected:    "param",
// 		},
// 		{
// 			name:        "fail",
// 			basePattern: "topic",
// 			testPattern: []byte("topic/paramA/paramB"),
// 			expected:    "",
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name,
// 			func(t *testing.T) {
// 				topic, err := NewTopic(tt.basePattern)
// 				if err != nil {
// 					t.Errorf("base pattern caused an error\nbase pattern: %s", tt.basePattern)
// 				}
//
// 				param := topic.paramMatcher(tt.testPattern)
//
// 				if param != tt.expected {
// 					t.Errorf("expected: %s\nreceived: %s", tt.expected, param)
// 				}
// 			},
// 		)
// 	}
// }
