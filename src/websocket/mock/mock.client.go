package domain_websocket_mock

import "context"

type mockClient struct {
	send chan []byte
}

func NewMockClient(send chan []byte) *mockClient {
	return &mockClient{send}
}

func (c *mockClient) Send(message []byte) {
	go func() {
		c.send <- message
	}()
}

func (c *mockClient) GetID() int {
	return 0
}

func (c *mockClient) ReadPump(ctx context.Context) {

}

func (c *mockClient) WritePump(ctx context.Context) {

}
