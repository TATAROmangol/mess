package transport

import (
	"github.com/TATAROmangol/mess/shared/logger"
	"github.com/TATAROmangol/mess/websocket/internal/loglables"
	"github.com/TATAROmangol/mess/websocket/internal/model"
)

type Hub struct {
	lg logger.Logger

	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client

	messageChan chan *model.Message
}

func NewHub(messageChan chan *model.Message, lg logger.Logger) *Hub {
	return &Hub{
		lg: lg,

		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),

		messageChan: messageChan,
	}
}

func (h *Hub) Run() {
	for {
		select {

		case client := <-h.register:
			h.clients[client.SubjectID] = client
			h.lg.With(loglables.Subject, client.SubjectID).Info("register")

		case client := <-h.unregister:
			if client, ok := h.clients[client.SubjectID]; ok {
				delete(h.clients, client.SubjectID)
				close(client.Send)
				h.lg.With(loglables.Subject, client.SubjectID).Info("unregister")
			}

		case message := <-h.messageChan:
			client, ok := h.clients[message.SubjectID]
			if !ok {
				continue
			}

			select {
			case client.Send <- message.WSMessage:
				h.lg.With(loglables.Subject, client.SubjectID).Info("send message")
			default:
				delete(h.clients, client.SubjectID)
				close(client.Send)
			}
		}
	}
}
