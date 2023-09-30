package domain_websocket_mock

import (
	"context"
	"encoding/json"
	"log"
)

type MockClient struct {
	send chan []byte
}

func NewMockClient(clientChan chan []byte) MockClient {
	return MockClient{
		clientChan,
	}
}

func (c MockClient) GetID() int {
	return 0
}

func (c MockClient) Send(message []byte) {
	c.send <- message
}

func (c MockClient) SendError(topic string, errorMsg string, logMsg string) error {
	error := make(map[string]string)
	error["topic"] = topic
	error["type"] = "error"
	error["payload"] = errorMsg

	jsonData, err := json.Marshal(error)

	if err != nil {
		log.Printf(logMsg, err)
		return err
	} else {
		go c.Send(jsonData)
		return nil
	}

}

func (c MockClient) ReadPump(
	ctx context.Context,
) {
}

func (c MockClient) WritePump(ctx context.Context) {
}
