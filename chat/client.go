package chat

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"time"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the chat.
type Client struct {
	ID int64

	Hub *Hub

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
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

// ReadPump pumps messages from the websocket connection to the chat.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
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

// WritePump pumps messages from the chat to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The chat closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			jsoned, _ := json.Marshal(message)

			w.Write(jsoned)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
