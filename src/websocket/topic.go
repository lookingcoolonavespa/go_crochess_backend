package domain_websocket

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	SubscribeEvent = "subscribeAndGet"
	InsertEvent    = "insert"
)

type TopicEventHandler = func(*Client, []byte) error

type Topic struct {
	match        func(bytes []byte) bool
	paramMatcher func(bytes []byte) string
	rooms        map[string]*Client
	events       map[string]TopicEventHandler
}

func NewTopic(
	// pattern follows the format: topic/param where param should be omitted
	// if there are no params
	pattern string,
) (*Topic, error) {
	patternRE, err := regexp.Compile("^[a-zA-Z0-9_]+(\\/[a-zA-Z0-9_]+)?$")
	if err != nil {
		return nil, err
	}

	if !patternRE.MatchString(pattern) {
		return nil, errors.New(fmt.Sprintf("pattern is not valid: %s", pattern))
	}

	patternSplit := strings.Split(pattern, "/")
	topic := patternSplit[0]
	topicRE, err := regexp.Compile(fmt.Sprintf("^%s+(\\/[a-zA-Z0-9_]+)?$", topic))
	if err != nil {
		return nil, err
	}

	matcher := func(bytes []byte) bool {
		return topicRE.Match(bytes)
	}

	paramRE, err := regexp.Compile("\\/\\w+$")
	if err != nil {
		return nil, err
	}

	paramMatcher := func(bytes []byte) string {
		base := topicRE.Find(bytes)
		if len(base) == 0 {
			return ""
		}
		return string(paramRE.Find(base))[1:]
	}

	return &Topic{
			matcher,
			paramMatcher,
			make(map[string]*Client),
			make(map[string]TopicEventHandler),
		},
		nil
}

func (t *Topic) HandleWSMessage(client *Client, event string, payload []byte) error {
	if handleFunc, ok := t.events[event]; ok {
		handleFunc(client, payload)
		return nil
	} else {
		return errors.New(fmt.Sprintf("%s is not a registered event", event))
	}
}

func (t *Topic) RegisterEvent(event string, handleFunc TopicEventHandler) {
	t.events[event] = handleFunc
}
