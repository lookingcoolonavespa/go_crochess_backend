package domain_websocket

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

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
		internalError := handleFunc(ctx, twp.room, client, payload)
		if internalError != nil {
			return errors.New("Something went wrong.")
		}
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
