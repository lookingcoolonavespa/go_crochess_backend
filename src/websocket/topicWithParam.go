package domain_websocket

import (
	"context"
	"fmt"
	"regexp"
)

type TopicWithParam struct {
	name      string
	matcher   *regexp.Regexp
	findParam func(string) string
	rooms     map[string]*Room
	events    map[string]TopicEventHandler
}

func (tp TopicWithParam) match(str string) bool {
	return tp.matcher.MatchString(str)
}

func (tp TopicWithParam) HandleWSMessage(
	ctx context.Context,
	client *Client,
	event string,
	payload []byte,
	topicName string,
) error {
	handleFunc, ok := tp.events[event]
	if !ok {
		err := client.SendError(
			fmt.Sprintf(`"%s" is not a registered event in topic "%s"`, event, tp.name),
			"TopicWithParam/HandleWSMessage: error transforming error message to json\nerr: %v",
		)
		return err

	}

	param := tp.findParam(topicName)
	room, ok := tp.rooms[param]
	if event == SubscribeEvent && !ok {
		room = NewRoom([]*Client{client}, param)
		err := tp.PushNewRoom(room)
		if err != nil {
			return err
		}
	}

	_, subscribed := room.clients[client.GetID()]
	if event != SubscribeEvent && !subscribed {
		err := client.SendError(
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

func (tp TopicWithParam) PushNewRoom(room *Room) error {
	param, err := room.GetParam()
	if err != nil {
		return err
	}

	tp.rooms[param] = room
	return nil
}
