package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"time"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	ID   int64
	Hub  *Hub
	Conn *websocket.Conn
	Send chan Message
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		ID:   time.Now().Unix(),
		Hub:  hub,
		Conn: conn,
		Send: make(chan Message, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		_ = c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		trimmedMsg := bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		msg := Message{
			Typing:   true,
			Message:  string(trimmedMsg[:]),
			ClientID: int(c.ID),
		}

		split := strings.Split(string(message[:]), "\n")
		if len(split) > 1 {
			msg.Typing = false

			c.Hub.RegisterHistory <- msg
		}

		c.Hub.Broadcast <- msg
	}
}

func (c *Client) WritePump() {
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			fmt.Printf("error in writepump: %v", err)
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The chat closed the channel.
				err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					fmt.Printf("error writing message: %v", err)
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			jsoned, _ := json.Marshal(message)

			_, err = w.Write(jsoned)
			if err != nil {
				fmt.Printf("error writing json data to writer: %v", err)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
