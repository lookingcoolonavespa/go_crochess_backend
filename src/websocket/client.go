package domain_websocket

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"nhooyr.io/websocket"
)

const (
	NormalCloseMessage = "web socket connection closing gracefully..."
)

type Client interface {
	GetID() int
	Subscribe(*Room)
	Unsubscribe(*Room)
	SendBytes([]byte)
	SendMessage(topic string, event string, payload interface{}, logMsg string) error
	ReadPump(context.Context)
	WritePump(context.Context)
	SendError(topic string, errorMsg string, logMsg string) error
}

type client struct {
	id       int
	conn     *websocket.Conn
	send     chan []byte
	wsServer *WebSocketServer
	rooms    map[*Room]bool
}

func NewClient(
	id int,
	sendChan chan []byte,
	conn *websocket.Conn,
	wsServer *WebSocketServer,
) Client {
	return &client{
		id,
		conn,
		sendChan,
		wsServer,
		make(map[*Room]bool, 0),
	}
}

func (c client) GetID() int {
	return c.id
}

func (c client) Subscribe(room *Room) {
	c.rooms[room] = true
	room.RegisterClient(c)
}

func (c client) Unsubscribe(room *Room) {
	if _, ok := c.rooms[room]; ok {
		delete(c.rooms, room)
	}
	room.RegisterClient(c)
}

func (c client) SendBytes(bytes []byte) {
	c.send <- bytes
}

func (c client) SendMessage(topic string, event string, payload interface{}, logFormat string) error {
	message, err := NewOutboundMessage(
		topic,
		event,
		payload,
	).ToJSON(logFormat)

	if err != nil {
		return err
	} else {
		go c.SendBytes(message)
		return nil
	}
}

func (c client) SendError(topic string, errorMsg string, logFormat string) error {
	return c.SendMessage(topic, ErrorEvent, errorMsg, logFormat)
}

func (c client) handleClose(ctx context.Context, err error) {
	if err == nil {
		return
	}

	defer func() {
		c.wsServer.unregister <- c
	}()

	if errors.Is(err, context.Canceled) {
		c.conn.Close(websocket.StatusNormalClosure, NormalCloseMessage)
		return
	} else if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		c.conn.Close(websocket.CloseStatus(err), NormalCloseMessage)
		return
	} else if err != nil {
		c.conn.Write(
			ctx,
			websocket.MessageText,
			[]byte(fmt.Sprintf("Something went wrong. Closing web socket connection. Error: %v", err)),
		)
		log.Printf("closing websock connection\nerror: %v", err)
		c.conn.Close(websocket.StatusInternalError, "")
		return
	}
}

func (c client) ReadPump(
	ctx context.Context,
) {
	for {
		_, _, err := c.conn.Reader(ctx)
		if err != nil {
			c.handleClose(ctx, err)
			break
		}
		buffer := make([]byte, 10000)
		if err != nil {
			c.handleClose(ctx, err)
			break
		}

		err = c.wsServer.router.HandleWSMessage(ctx, c, buffer)
		if err != nil {
			c.handleClose(ctx, err)
			break
		}
	}
}

func (c client) WritePump(ctx context.Context) {
	pingTimer := time.NewTicker(PingPeriod)
	defer func() {
		pingTimer.Stop()
	}()
	for {
		select {
		case message := <-c.send:
			w, err := c.conn.Writer(ctx, websocket.MessageText)
			if err != nil {
				c.handleClose(ctx, err)
				return
			}

			w.Write(message)

			if err := w.Close(); err != nil {
				c.handleClose(ctx, err)
				return
			}

		case <-pingTimer.C:
			err := c.conn.Ping(ctx)
			if err != nil {
				c.handleClose(ctx, err)
				return
			}
		}
	}
}
