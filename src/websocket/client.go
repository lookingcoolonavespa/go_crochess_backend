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

type Client struct {
	conn     *websocket.Conn
	Send     chan []byte
	wsServer *WebSocketServer
}

func NewClient(conn *websocket.Conn, wsServer *WebSocketServer) *Client {
	return &Client{
		conn,
		make(chan []byte, 256),
		wsServer,
	}
}

func (c *Client) handleClose(ctx context.Context, err error) {
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

func (c *Client) readPump(
	ctx context.Context,
) {
	for {
		_, wsMessage, err := c.conn.Reader(ctx)
		if err != nil {
			c.handleClose(ctx, err)
			break
		}
		buffer := make([]byte, 10000)
		bufferLen, err := wsMessage.Read(buffer)
		if err != nil {
			c.handleClose(ctx, err)
			break
		}

		err = c.wsServer.router.HandleWSMessage(ctx, c, buffer, bufferLen)
		if err != nil {
			c.handleClose(ctx, err)
			break
		}
	}
}

func (c *Client) writePump(ctx context.Context) {
	pingTimer := time.NewTicker(PingPeriod)
	defer func() {
		pingTimer.Stop()
	}()
	for {
		select {
		case message := <-c.Send:
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
