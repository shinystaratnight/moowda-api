package sockets

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"

	"moowda/models"
)

// Hub class
type Hub struct {
	clients    map[*Client]bool
	topicsCh   chan *models.TopicDetail
	messagesCh chan *models.TopicMessage
	register   chan *Client
	unregister chan *Client
	topics     map[int][]*Client
}

// NewHub func
func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		topicsCh:   make(chan *models.TopicDetail),
		messagesCh: make(chan *models.TopicMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		topics:     map[int][]*Client{},
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
		case topic := <-h.topicsCh:
			messageType := "topic_created"
			if topic.MessagesCount > 0 {
				messageType = "‘topic_message_added"
			}

			type topicMessage struct {
				Type  string              `json:"type"`
				Topic *models.TopicDetail `json:"topic"`
			}

			resp := &topicMessage{
				Type:  messageType,
				Topic: topic,
			}

			data, err := json.Marshal(resp)
			if err == nil {
				h.send(data)
			}
		case msg := <-h.messagesCh:
			fmt.Printf("read %v", msg.Content)
			messageType := "message_added"

			type message struct {
				Type    string               `json:"type"`
				Message *models.TopicMessage `json:"message"`
			}

			resp := &message{
				Type:    messageType,
				Message: msg,
			}

			data, err := json.Marshal(resp)
			if err == nil {
				fmt.Printf("send %v", msg.Topic.ID)
				h.sendToTopic(int(msg.Topic.ID), data)
			}
		}
	}
}

func (h *Hub) sendToTopic(topicID int, data []byte) {
	for _, client := range h.topics[topicID] {
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(h.clients, client)
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

func (h *Hub) BroadcastTopic(topic *models.TopicDetail) {
	h.topicsCh <- topic
}

func (h *Hub) BroadcastMessage(message *models.TopicMessage) {
	h.messagesCh <- message
}

// RunHub func
func RunTopicsHub(e *echo.Echo) *Hub {
	hub := newHub()
	go hub.Run()

	e.GET("/ws/topics/events", func(c echo.Context) (err error) {
		err = serveWs(hub, c)
		return
	})

	return hub
}

// RunHub func
func RunMessagesHub(e *echo.Echo) *Hub {
	hub := newHub()
	go hub.Run()

	e.GET("/ws/topics/:id/events", func(c echo.Context) (err error) {
		err = serveWs(hub, c)
		return
	})

	return hub
}
