package chat

type Message struct {
	Typing   bool   `json:"typing"`
	Message  string `json:"message"`
	ClientID int    `json:"client_id"`
}

type Hub struct {
	Clients         map[*Client]bool
	Broadcast       chan Message
	Register        chan *Client
	Unregister      chan *Client
	MsgHistory      []Message
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
				client.Send <- msg
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
