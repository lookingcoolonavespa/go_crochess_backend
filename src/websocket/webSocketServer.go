package domain_websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

const (
	PingPeriod = time.Second * 30
)

type WebSocketServer struct {
	conns      map[*Client]bool
	register   chan *Client
	unregister chan *Client
	router     WebSocketRouter
}

func NewWebSocketServer(r WebSocketRouter) *WebSocketServer {
	return &WebSocketServer{
		conns:      make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		router:     r,
	}
}

func (s *WebSocketServer) Run() {
	select {
	case client := <-s.unregister:
		s.unregisterClient(client)
	}
}

func (s *WebSocketServer) HandleWS(w http.ResponseWriter, r *http.Request) {
	wsConfig := websocket.AcceptOptions{
		InsecureSkipVerify: false,
		OriginPatterns:     []string{"*"},
		CompressionMode:    websocket.CompressionDisabled,
	}
	conn, err := websocket.Accept(w, r, &wsConfig)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	client := NewClient(conn, s)

	go client.readPump(r.Context())
	go client.writePump(r.Context())

}

func (s *WebSocketServer) registerClient(client *Client) {
	s.conns[client] = true
}

func (s *WebSocketServer) unregisterClient(client *Client) {
	if _, ok := s.conns[client]; ok {
		delete(s.conns, client)
	}
}

func (s *WebSocketServer) HandleWSMessage(ctx context.Context, client *Client, message []byte, messageLength int) error {
	var wsMessage WebSocketMessage
	err := json.Unmarshal(message[:messageLength], &wsMessage)
	if err != nil {
		log.Println("something went wrong reading a message: ", err)
		return err
	}

	return nil
}
