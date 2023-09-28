package domain_websocket

import "encoding/json"

type InboundMessage struct {
	Room    string
	Event   string
	payload []byte
}

type OutboundMessage struct {
	Topic   string
	Event   string
	payload interface{}
}

func NewOutboundMessage(topic string, event string, payload interface{}) *OutboundMessage {
	return &OutboundMessage{
		topic,
		event,
		payload,
	}
}

func (m *OutboundMessage) ToJSON() ([]byte, error) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}
