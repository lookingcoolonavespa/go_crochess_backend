package domain_websocket

import (
	"errors"
	"fmt"
)

type Room struct {
	clients map[int]Client
	param   string
}

func NewRoom(clients []Client, param string) Room {
	var clientMap map[int]Client
	for _, client := range clients {
		clientMap[client.GetID()] = client
	}
	return Room{
		clientMap,
		param,
	}
}

func (r *Room) BroadcastMessage(message []byte) {
	for _, client := range r.clients {
		client.Send(message)
	}
}

func (r *Room) PushNewClient(client Client) error {
	_, ok := r.clients[client.GetID()]
	if ok {
		return errors.New(fmt.Sprintf(`a client with the id "%d" already exists`, client.GetID()))
	}

	r.clients[client.GetID()] = client

	return nil
}

func (r *Room) GetParam() (string, error) {
	if r.param == "" {
		return "", errors.New("room does not have a param")
	}

	return r.param, nil
}
