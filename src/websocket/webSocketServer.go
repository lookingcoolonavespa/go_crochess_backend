package domain_websocket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"nhooyr.io/websocket"
)

const (
	PingPeriod = time.Second * 30
)

var clientID = 0

type WebSocketServer struct {
	conns         map[*Client]bool
	router        WebSocketRouter
	mutex         sync.Mutex
	gameseeksRepo domain.GameseeksRepo
}

func NewWebSocketServer(r WebSocketRouter, gameseeksRepo domain.GameseeksRepo) WebSocketServer {
	return WebSocketServer{
		conns:         make(map[*Client]bool),
		router:        r,
		gameseeksRepo: gameseeksRepo,
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
		log.Printf("%v", err)
		return
	}

	var userID int
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		userID = clientID
		clientID += 1
	} else {
		userID, err = strconv.Atoi(uid)
		if err != nil {
			log.Printf("error parsing user id when connecting: %v", err)
			conn.Close(websocket.StatusInternalError, fmt.Sprintf(`"%s" is not a valid uid`, uid))
			return
		}
	}
	client := NewClient(userID, make(chan []byte), conn, s)
	log.Println("client connected: ", userID)

	go client.ReadPump(r.Context())
	go client.WritePump(r.Context())

	s.registerClient(client)

	select {
	case <-r.Context().Done():
		{
			fmt.Println("client disconnected")
		}
	}
}

func (s *WebSocketServer) registerClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.conns[client] = true
}

func (s *WebSocketServer) unregisterClient(ctx context.Context, client *Client) {
	s.mutex.Lock()
	if _, ok := s.conns[client]; ok {
		delete(s.conns, client)
	}
	s.mutex.Unlock()

	log.Println("unregistering client: ", client.GetID())
	s.gameseeksRepo.DeleteFromSeeker(ctx, client.GetID())
}

func (s *WebSocketServer) Close() {
	for client := range s.conns {
		client.HandleClose(context.Background(), context.Canceled)
	}
}
