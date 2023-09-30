package domain_websocket

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

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

	internalErr := handleFunc(ctx, room, client, payload)
	if internalErr != nil {
		return errors.New("Something went wrong.")
	}

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
