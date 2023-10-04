package domain_websocket

import (
	"errors"
	"fmt"
	"sync"
)

type Room struct {
	clients map[int]*Client
	param   string
	mutex   sync.Mutex
}

func NewRoom(clients []*Client, param string) *Room {
	clientMap := make(map[int]*Client)
	for _, client := range clients {
		clientMap[client.GetID()] = client
	}
	return &Room{
		clientMap,
		param,
		sync.Mutex{},
	}
}

func (r *Room) BroadcastMessage(message []byte) {
	for _, client := range r.clients {
		go client.SendBytes(message)
	}
}

func (r *Room) RegisterClient(client *Client) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, ok := r.clients[client.GetID()]
	if ok {
		return errors.New(fmt.Sprintf(`a client with the id "%d" already exists`, client.GetID()))
	}

	r.clients[client.GetID()] = client

	return nil
}

func (r *Room) UnregisterClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.clients[client.GetID()]; ok {
		delete(r.clients, client.GetID())
	}
}

func (r *Room) GetParam() (string, error) {
	if r.param == "" {
		return "", errors.New("room does not have a param")
	}

	return r.param, nil
}
