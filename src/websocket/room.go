package domain_websocket

import (
	"errors"
	"fmt"
	"sync"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
)

type Room struct {
	clients map[string]domain.Client
	param   string
	mutex   sync.Mutex
}

func NewRoom(clients []domain.Client, param string) *Room {
	clientMap := make(map[string]domain.Client)
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
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, client := range r.clients {
		go client.SendBytes(message)
	}
}

func (r *Room) RegisterClient(client domain.Client) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, ok := r.clients[client.GetID()]
	if ok {
		return errors.New(fmt.Sprintf(`a client with the id "%s" already exists`, client.GetID()))
	}

	r.clients[client.GetID()] = client

	return nil
}

func (r *Room) UnregisterClient(client domain.Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.clients[client.GetID()]; ok {
		delete(r.clients, client.GetID())
	}
}

func (r *Room) ChangeParam(param string) {
	r.param = param
}

func (r *Room) GetParam() (string, error) {
	if r.param == "" {
		return "", errors.New("room does not have a param")
	}

	return r.param, nil
}

func (r *Room) GetClient(id string) (domain.Client, bool) {
	client, ok := r.clients[id]
	return client, ok
}
