package domain_websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type WebSocketRouter struct {
	topics []Topic
}

func NewWebSocketRouter() (WebSocketRouter, error) {

	return WebSocketRouter{
			make([]Topic, 0),
		},
		nil
}

func (r *WebSocketRouter) PushNewRoute(topic Topic) {
	r.topics = append(r.topics, topic)
}

func (r WebSocketRouter) HandleWSMessage(ctx context.Context, client *Client, jsonMessage []byte) error {
	var message InboundMessage
	err := json.Unmarshal(jsonMessage, &message)
	if err != nil {
		log.Printf("WebSocketRouter/HandleWSMessage: error unmarshalling message\nerr: %v", err)
	}

	filled, missingFields := message.IsFilled()
	if !filled {
		err := client.SendError(
			fmt.Sprint("message is missing fields: ",
				strings.Join(missingFields, ", "),
			),
			"WebSocketRouter/HandleWSMessage: error transforming error message to json\nerr: %v",
		)
		return err
	}

	for _, topic := range r.topics {
		if topic.match(message.Topic) {
			internalErr := topic.HandleWSMessage(
				ctx,
				client,
				message.Event,
				message.Payload,
				message.Topic,
			)
			return internalErr
		}
	}

	err = client.SendError(
		fmt.Sprintf(`"%s" is not a valid topic`, message.Topic),
		"WebSocketRouter/HandleWSMessage: error transforming error message to json\nerr: %v",
	)
	return err
}
