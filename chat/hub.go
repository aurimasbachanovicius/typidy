package chat

type Message struct {
	Typing   bool   `json:"typing"`
	Message  string `json:"message"`
	ClientID int    `json:"client_id"`
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	Clients map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan Message

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	// MsgHistory stores all inbounded messages
	MsgHistory []Message

	// RegisterHistory channel for registering message in history
	RegisterHistory chan Message
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:       make(chan Message),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		Clients:         make(map[*Client]bool),
		MsgHistory:      []Message{},
		RegisterHistory: make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			for _, msg := range h.MsgHistory {
				select {
				case client.Send <- msg:
				default:
					close(client.Send)
				}
			}

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}

		case message := <-h.RegisterHistory:
			h.MsgHistory = append(h.MsgHistory, message)

		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
