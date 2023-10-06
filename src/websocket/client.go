package domain_websocket

import (
	"context"
	"errors"
	"fmt"
	"log"

	"nhooyr.io/websocket"
)

const (
	NormalCloseMessage = "web socket connection closing gracefully..."
)

type Client struct {
	id       string
	conn     *websocket.Conn
	send     chan []byte
	wsServer *WebSocketServer
	rooms    map[*Room]bool
}

func NewClient(
	id string,
	sendChan chan []byte,
	conn *websocket.Conn,
	wsServer *WebSocketServer,
) *Client {
	return &Client{
		id,
		conn,
		sendChan,
		wsServer,
		make(map[*Room]bool, 0),
	}
}

func (c Client) GetID() string {
	return c.id
}

func (c *Client) Subscribe(room *Room) error {
	err := room.RegisterClient(c)
	if err != nil {
		return err
	}
	_, subscribed := c.rooms[room]
	if subscribed {
		return errors.New("client already has this subscription")
	}

	c.rooms[room] = true

	return nil
}

func (c *Client) Unsubscribe(room *Room) {
	if _, ok := c.rooms[room]; ok {
		delete(c.rooms, room)
	}
	room.UnregisterClient(c)
}

func (c *Client) SendBytes(bytes []byte) {
	c.send <- bytes
}

func (c *Client) SendMessage(topic string, event string, payload interface{}, logFormat string) error {
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

func (c *Client) SendError(errorMsg string, logFormat string) error {
	return c.SendMessage(ErrorEvent, ErrorEvent, errorMsg, logFormat)
}

func (c *Client) HandleClose(ctx context.Context, err error) {
	if err == nil {
		return
	}

	defer func() {
		c.wsServer.unregisterClient(ctx, c)
		for room := range c.rooms {
			if room != nil {
				c.Unsubscribe(room)
			}
		}
	}()

	if errors.Is(err, context.Canceled) {
		c.conn.Close(websocket.StatusNormalClosure, NormalCloseMessage)
		return
	} else if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		c.conn.Close(websocket.CloseStatus(err), NormalCloseMessage)
		return
	} else if err != nil {
		defer func() {
			c.conn.Close(websocket.StatusInternalError, "")
		}()

		log.Printf("closing websocket connection, error: %v", err)

		errorMessage, _ := NewOutboundMessage(
			"error",
			ErrorEvent,
			fmt.Sprintf("Something went wrong. Closing web socket connection. Error: %v", err),
		).ToJSON("Client/HandleClose, error converting error messsage to json: %v")

		c.conn.Write(
			ctx,
			websocket.MessageText,
			errorMessage,
		)
		return
	}
}

func (c *Client) ReadPump(
	ctx context.Context,
) {
	for {
		_, r, err := c.conn.Reader(ctx)
		if err != nil {
			c.HandleClose(ctx, err)
			break
		}
		buffer := make([]byte, 10000)
		messageLen, err := r.Read(buffer)
		if err != nil {
			c.HandleClose(ctx, err)
			break
		}

		err = c.wsServer.router.HandleWSMessage(ctx, c, buffer[0:messageLen])
		if err != nil {
			c.HandleClose(ctx, err)
			break
		}
	}
}

func (c *Client) WritePump(ctx context.Context) {
	for {
		select {
		case message := <-c.send:
			w, err := c.conn.Writer(ctx, websocket.MessageText)
			if err != nil {
				c.HandleClose(ctx, err)
				return
			}

			w.Write(message)

			if err := w.Close(); err != nil {
				c.HandleClose(ctx, err)
				return
			}
		}
	}
}
