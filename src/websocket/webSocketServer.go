package domain_websocket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/spf13/viper"
	"nhooyr.io/websocket"
)

const (
	PingPeriod = time.Second * 30
)

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

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", viper.GetString("production.origin"))
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next(w, r)
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

	uid := r.URL.Query().Get("uid")
	client := NewClient(uid, make(chan []byte), conn, s)
	log.Println("client connected: ", uid)

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
