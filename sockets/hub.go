package sockets

import (
	"encoding/json"
	"github.com/labstack/echo"

	"moowda/models"
)

// Hub class
type Hub struct {
	clients    map[*Client]bool
	topics     chan *models.TopicCard
	register   chan *Client
	unregister chan *Client
}

// NewHub func
func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		topics:     make(chan *models.TopicCard),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run func
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case topic := <-h.topics:
			type topicMessage struct {
				Type  string            `json:"type"`
				Topic *models.TopicCard `json:"topic"`
			}

			resp := &topicMessage{
				Type:  "topic_created",
				Topic: topic,
			}

			data, err := json.Marshal(resp)
			if err == nil {
				h.send(data)
			}
		}
	}
}

func (h *Hub) send(data []byte) {
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) BroadcastTopic(topic *models.TopicCard) {
	h.topics <- topic
}

// RunHub func
func RunHub(e *echo.Echo) *Hub {
	hub := newHub()
	go hub.Run()

	e.GET("/ws/topics/events", func(c echo.Context) (err error) {
		err = serveWs(hub, c)
		return
	})

	return hub
}
