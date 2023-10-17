package domain

import "context"

type Room interface {
	BroadcastMessage(message []byte)
	RegisterClient(client Client) error
	UnregisterClient(client Client)
	ChangeParam(param string)
	GetParam() (string, error)
	GetClient(id string) (Client, bool)
}

type Client interface {
	GetID() string
	Subscribe(room Room) error
	Unsubscribe(room Room)
	SendBytes(bytes []byte)
	SendMessage(topic string, event string, payload interface{}, logFormat string) error
	SendError(errorMsg string, logFormat string) error
	HandleClose(ctx context.Context, err error)
	ReadPump(ctx context.Context)
	WritePump(ctx context.Context)
}
