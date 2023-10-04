package domain_websocket

import (
	"encoding/json"
	"log"
)

type InboundMessage struct {
	Topic   string          `json:"topic"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

func (m InboundMessage) IsFilled() (bool, []string) {
	missingFields := make([]string, 0)
	if m.Topic == "" {
		missingFields = append(missingFields, "topic")
	}
	if m.Event == "" {
		missingFields = append(missingFields, "event")
	}

	return len(missingFields) == 0, missingFields
}

type OutboundMessage struct {
	Topic   string      `json:"topic"`
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

func NewOutboundMessage(topic string, event string, payload interface{}) OutboundMessage {
	return OutboundMessage{
		topic,
		event,
		payload,
	}
}

func (m OutboundMessage) ToJSON(logFormat string) ([]byte, error) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		log.Printf(logFormat, err)
		return nil, err
	}

	return jsonData, nil
}
