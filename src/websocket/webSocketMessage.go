package domain_websocket

const (
	SubscribeAndGet = "subscribeGet"
)

type WebSocketMessage struct {
	Room    string
	Event   string
	payload []byte
}
