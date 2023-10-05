package domain_websocket

import (
	"context"
	"fmt"
	"regexp"
)

type TopicWithoutParm struct {
	name    string
	matcher *regexp.Regexp
	room    *Room
	events  map[string]TopicEventHandler
}

func (twp TopicWithoutParm) match(str string) bool {
	return twp.matcher.MatchString(str)
}

func (twp TopicWithoutParm) HandleWSMessage(
	ctx context.Context,
	client *Client,
	event string,
	payload []byte,
	_ string,
) error {
	handleFunc, ok := twp.events[event]
	if !ok {
		err := client.SendError(
			fmt.Sprintf(`"%s" is not a registered event in topic "%s"`, event, twp.name),
			"TopicWithoutParam/HandleWSMessage: error transforming error message to json\nerr: %v",
		)
		return err
	}

	_, subscribed := twp.GetClient(client.GetID())
	if event != SubscribeEvent && !subscribed {
		err := client.SendError(
			fmt.Sprintf(`you are not subscribed to "%s"`, twp.name),
			"TopicWithoutParam/HandleWSMessage: error transforming error message to json\nerr: %v",
		)
		return err
	}

	internalError := handleFunc(ctx, twp.room, client, payload)
	return internalError
}

func (twp TopicWithoutParm) RegisterEvent(event string, handleFunc TopicEventHandler) {
	twp.events[event] = handleFunc
}

func (twp TopicWithoutParm) GetClient(id string) (*Client, bool) {
	client, ok := twp.room.clients[id]
	return client, ok
}
