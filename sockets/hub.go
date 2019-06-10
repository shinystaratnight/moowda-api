package sockets

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"moowda/services"

	"moowda/models"
)

// Hub class
type Hub struct {
	clients        map[*Client]bool
	topicUpdatedCh chan *models.Topic
	messagesCh     chan *models.TopicMessage
	register       chan *Client
	unregister     chan *Client
	topics         map[int][]*Client
	topicService   *services.TopicService
}

// NewHub func
func newHub(topicService *services.TopicService) *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		topicUpdatedCh: make(chan *models.Topic),
		messagesCh:     make(chan *models.TopicMessage),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		topics:         map[int][]*Client{},
		topicService:   topicService,
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
		case topic := <-h.topicUpdatedCh:
			type topicMessage struct {
				Type  string            `json:"type"`
				Topic *models.TopicCard `json:"topic"`
			}

			for client := range h.clients {
				topicCard, err := h.topicService.GetTopicCardForUser(topic, client.user)
				if err != nil {
					log.Error(err)
					continue
				}

				messageType := "topic_created"
				if topicCard.MessagesCount > 0 {
					messageType = "topic_message_added"
				}
				resp := &topicMessage{
					Type:  messageType,
					Topic: topicCard,
				}

				data, err := json.Marshal(resp)
				if err != nil {
					log.Error(err)
					continue
				}
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
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

func (h *Hub) BroadcastTopic(topic *models.Topic) {
	h.topicUpdatedCh <- topic
}

func (h *Hub) BroadcastMessage(message *models.TopicMessage) {
	h.messagesCh <- message
}

// RunHub func
func RunTopicsHub(e *echo.Echo, topicService *services.TopicService) *Hub {
	hub := newHub(topicService)
	go hub.Run()

	e.GET("/ws/topics/events", func(c echo.Context) (err error) {
		err = serveWs(hub, c)
		return
	})

	return hub
}

// RunHub func
func RunMessagesHub(e *echo.Echo, topicService *services.TopicService) *Hub {
	hub := newHub(topicService)
	go hub.Run()

	e.GET("/ws/topics/:id/events", func(c echo.Context) (err error) {
		err = serveWs(hub, c)
		return
	})

	return hub
}
