package domain_websocket

import (
	"context"
	"fmt"
	"regexp"
)

type WebSocketRouter struct {
	topics       []Topic
	topicRegex   *regexp.Regexp
	eventRegex   *regexp.Regexp
	payloadRegex *regexp.Regexp
}

func jsonRegex(field string) (*regexp.Regexp, error) {
	pattern := fmt.Sprintf(`"%s"\s*:\s*"([^"]+)"`, field)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return re, nil
}

func NewWebSocketRouter() (WebSocketRouter, error) {
	topicRegex, err := jsonRegex("topic")
	if err != nil {
		return WebSocketRouter{}, err
	}

	eventRegex, err := jsonRegex("event")
	if err != nil {
		return WebSocketRouter{}, err
	}

	payloadRegex, err := jsonRegex("payload")
	if err != nil {
		return WebSocketRouter{}, err
	}

	return WebSocketRouter{
			make([]Topic, 0),
			topicRegex,
			eventRegex,
			payloadRegex,
		},
		nil
}

func (r *WebSocketRouter) PushNewRoute(topic Topic) {
	r.topics = append(r.topics, topic)
}

func (r WebSocketRouter) HandleWSMessage(ctx context.Context, client Client, message []byte) error {
	topicBytes := r.topicRegex.FindSubmatch(message)
	if len(topicBytes) != 2 {
		err := client.SendError(
			"*",
			"message is not in the correct format: missing topic field",
			"WebSocketRouter/HandleWSMessage: error transforming error message to json\nerr: %v",
		)
		return err
	}

	for _, topic := range r.topics {
		if topic.match(topicBytes[1]) {
			eventBytes := r.eventRegex.FindSubmatch(message)
			if len(eventBytes) != 2 {
				err := client.SendError(
					"*",
					"message is not in the correct format: missing event field",
					"WebSocketRouter/HandleWSMessage: error transforming error message to json\nerr: %v",
				)
				return err
			}
			payloadBytes := r.payloadRegex.FindSubmatch(message)
			if len(payloadBytes) != 2 {
				err := client.SendError(
					"*",
					"message is not in the correct format: missing payload field",
					"WebSocketRouter/HandleWSMessage: error transforming error message to json\nerr: %v",
				)
				return err
			}

			internalErr := topic.HandleWSMessage(
				ctx,
				client,
				string(eventBytes[1]),
				payloadBytes[1],
				topicBytes[1],
			)
			return internalErr
		}
	}

	err := client.SendError(
		"*",
		fmt.Sprintf(`"%s" is not a valid topic`, string(topicBytes[1])),
		"WebSocketRouter/HandleWSMessage: error transforming error message to json\nerr: %v",
	)
	return err
}
