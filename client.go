package main

import (
	"log"
	"net"
	"time"

	"github.com/gobwas/ws/wsutil"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn net.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// ReadMessage pumps messages from the websocket connection to the hub.
//
// The application runs ReadMessage in a per-connection goroutine.
func (c *Client) ReadMessage() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	for {
		message, _, err := wsutil.ReadClientData(c.conn)
		if err != nil {
			log.Printf("Error: %v", err)
			break
		}
		c.hub.broadcast <- message
	}
}

// WriteMessage pumps messages from the hub to the websocket connection.
//
// A goroutine running WriteMessage is started for each connection.
func (c *Client) WriteMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				return
			}
			wsutil.WriteServerBinary(c.conn, message)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsutil.WriteServerBinary(c.conn, nil); err != nil {
				return
			}
		}
	}
}
