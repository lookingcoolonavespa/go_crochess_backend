package domain_websocket

import (
	"context"
	"fmt"
	"regexp"
)

type TopicWithParam struct {
	name      string
	matcher   *regexp.Regexp
	findParam func(bytes []byte) string
	rooms     map[string]*Room
	events    map[string]TopicEventHandler
}

func (tp TopicWithParam) match(bytes []byte) bool {
	return tp.matcher.Match(bytes)
}

func (tp TopicWithParam) HandleWSMessage(
	ctx context.Context,
	client Client,
	event string,
	payload []byte,
	topicName []byte,
) error {
	handleFunc, ok := tp.events[event]
	if !ok {
		err := client.SendError(
			tp.name,
			fmt.Sprintf("%s is not a registered event", event),
			"TopicWithParam/HandleWSMessage: error transforming error message to json\nerr: %v",
		)
		return err

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
		err := client.SendError(
			tp.name,
			fmt.Sprintf(`you are not subscribed to "%s/%s"`, tp.name, param),
			"TopicWithParam/HandleWSMessage: error transforming error message to json\nerr: %v",
		)
		return err
	}

	internalErr := handleFunc(ctx, room, client, payload)
	return internalErr

}

func (tp TopicWithParam) RegisterEvent(event string, handleFunc TopicEventHandler) {
	tp.events[event] = handleFunc
}

func (tp TopicWithParam) PushNewRoom(param string, clients []Client) *Room {
	room := NewRoom(clients, param)
	tp.rooms[param] = room
	return room
}
