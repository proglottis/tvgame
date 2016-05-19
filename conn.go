package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type ConnMessage struct {
	Type string
	Data json.RawMessage `json:",omitempty"`
	err  error
}

type Conn struct {
	ws     *websocket.Conn
	send   chan ConnMessage
	cancel context.CancelFunc
}

func NewConn(ctx context.Context, ws *websocket.Conn) *Conn {
	conn := &Conn{
		ws:   ws,
		send: make(chan ConnMessage, 500),
	}
	ctx, conn.cancel = context.WithCancel(ctx)
	conn.ws.SetReadLimit(maxMessageSize)
	conn.ws.SetReadDeadline(time.Now().Add(pongWait))
	conn.ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	go conn.writePump(ctx)
	return conn
}

func (c *Conn) Close() error {
	c.cancel()
	return nil
}

func (c *Conn) Read(msg *ConnMessage) error {
	if err := c.ws.ReadJSON(msg); err != nil {
		return err
	}
	return nil
}

func (c *Conn) Write(msg *ConnMessage) error {
	c.send <- *msg
	return nil
}

func (c *Conn) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.SetWriteDeadline(time.Now().Add(writeWait))
		c.ws.WriteMessage(websocket.CloseMessage, []byte{})
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteJSON(&message); err != nil {
				log.Printf("Conn: write: %s", err)
				return
			}
		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("Conn: ping: %s", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
