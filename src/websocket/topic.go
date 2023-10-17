package domain_websocket

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
)

type TopicEventHandler = func(context.Context, domain.Room, domain.Client, []byte) error

type Topic interface {
	HandleWSMessage(
		ctx context.Context,
		client *Client,
		event string,
		payload []byte,
		topicName string,
	) error
	RegisterEvent(event string, handleFunc TopicEventHandler)
	match(string) bool
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

	var topicRE *regexp.Regexp
	var paramMatcher func(string) string
	if len(patternSplit) > 1 {
		topicRE, err = regexp.Compile(fmt.Sprintf("^%s+(\\/[a-zA-Z0-9_]+)?$", topic))
		if err != nil {
			return nil, err
		}

		paramRE, err := regexp.Compile("\\/\\w+$")
		if err != nil {
			return nil, err
		}

		paramMatcher = func(test string) string {
			base := topicRE.FindString(test)
			if len(base) == 0 {
				return ""
			}
			return string(paramRE.FindString(base))[1:]
		}
	} else {
		topicRE, err = regexp.Compile(fmt.Sprintf("^%s$", topic))
		if err != nil {
			return nil, err
		}
	}

	if len(patternSplit) > 1 {
		return TopicWithParam{
				name:      topic,
				matcher:   topicRE,
				findParam: paramMatcher,
				rooms:     make(map[string]*Room),
				events:    make(map[string]TopicEventHandler),
			},
			nil
	} else {
		return TopicWithoutParm{
				name:    topic,
				matcher: topicRE,
				room:    NewRoom(make([]domain.Client, 0), ""),
				events:  make(map[string]TopicEventHandler),
			},
			nil
	}

}
