package domain_websocket

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type TopicEventHandler = func(context.Context, Room, Client, []byte) error

type Topic interface {
	HandleWSMessage(
		ctx context.Context,
		client Client,
		event string,
		payload []byte,
		topicName []byte,
	) error
	RegisterEvent(event string, handleFunc TopicEventHandler)
	match(bytes []byte) bool
}

func NewTopic(
	// pattern follows the format: topic/param where param should be omitted
	// if there are no params
	pattern string,
) (Topic, error) {
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

	paramRE, err := regexp.Compile("\\/\\w+$")
	if err != nil {
		return nil, err
	}

	var paramMatcher func(bytes []byte) string
	if len(patternSplit) > 1 {
		paramMatcher = func(bytes []byte) string {
			base := topicRE.Find(bytes)
			if len(base) == 0 {
				return ""
			}
			return string(paramRE.Find(base))[1:]
		}
	}

	if len(patternSplit) > 1 {
		return TopicWithParm{
				matcher:   topicRE,
				findParam: paramMatcher,
				rooms:     make(map[string]Room),
				events:    make(map[string]TopicEventHandler),
			},
			nil
	} else {
		return TopicWithoutParm{
				matcher: topicRE,
				room:    NewRoom(make([]Client, 0), ""),
				events:  make(map[string]TopicEventHandler),
			},
			nil
	}

}
