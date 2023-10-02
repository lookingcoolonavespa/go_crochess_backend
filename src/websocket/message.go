package domain_websocket

import (
	"encoding/json"
	"log"
)

type InboundMessage struct {
	Room    string
	Event   string
	payload []byte
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
