package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"luckydraw/internal/models"
)

type Hub struct {
	clients    map[*websocket.Conn]bool
	Broadcast  chan *models.Message
	Register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		Broadcast:  make(chan *models.Message),
		Register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			
		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) broadcastMessage(message *models.Message) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	
	h.mu.Lock()
	for client := range h.clients {
		err := client.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			client.Close()
			delete(h.clients, client)
		}
	}
	h.mu.Unlock()
}