package domain_websocket

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	SubscribeEvent = "subscribe"
	InsertEvent    = "insert"
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
				room:    NewRoom(make([]Client, 1), ""),
				events:  make(map[string]TopicEventHandler),
			},
			nil
	}

}

type TopicWithParm struct {
	matcher   *regexp.Regexp
	findParam func(bytes []byte) string
	rooms     map[string]Room
	events    map[string]TopicEventHandler
}

func (tp TopicWithParm) match(bytes []byte) bool {
	return tp.matcher.Match(bytes)
}

func (tp TopicWithParm) HandleWSMessage(ctx context.Context, client Client, event string, payload []byte, topicName []byte) error {
	handleFunc, ok := tp.events[event]
	if !ok {
		return errors.New(fmt.Sprintf("%s is not a registered event", event))
	}

	param := tp.findParam(topicName)
	room, ok := tp.rooms[param]
	if event == SubscribeEvent {
		if ok {
			room.PushNewClient(client)
		} else {
			room = tp.PushNewRoom(param, []Client{client})
		}
	}

	if event != SubscribeEvent && !ok {
		return errors.New(fmt.Sprintf(`could not find a room for param "%s"`, param))
	}

	handleFunc(ctx, room, client, payload)
	return nil
}

func (tp TopicWithParm) RegisterEvent(event string, handleFunc TopicEventHandler) {
	tp.events[event] = handleFunc
}

func (tp TopicWithParm) PushNewRoom(param string, clients []Client) Room {
	room := NewRoom(clients, param)
	tp.rooms[param] = room
	return room
}

type TopicWithoutParm struct {
	matcher *regexp.Regexp
	room    Room
	events  map[string]TopicEventHandler
}

func (twp TopicWithoutParm) match(bytes []byte) bool {
	return twp.matcher.Match(bytes)
}

func (twp TopicWithoutParm) HandleWSMessage(ctx context.Context, client Client, event string, payload []byte, _ []byte) error {
	if handleFunc, ok := twp.events[event]; ok {
		if event == SubscribeEvent {
			twp.room.PushNewClient(client)
		}
		handleFunc(ctx, twp.room, client, payload)
		return nil
	} else {
		return errors.New(fmt.Sprintf("%s is not a registered event", event))
	}
}

func (twp TopicWithoutParm) RegisterEvent(event string, handleFunc TopicEventHandler) {
	twp.events[event] = handleFunc
}

func (twp TopicWithoutParm) GetClient(id int) (Client, bool) {
	client, ok := twp.room.clients[id]
	return client, ok
}
