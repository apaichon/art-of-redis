package server

import (
    "log"
    "net/http"
    "sync"
	"context" 

    "github.com/gorilla/websocket"
    "leaderboard/internal/domain/events"
    "leaderboard/internal/domain/models"
    "leaderboard/internal/ports"
)

type WebSocketHub struct {
    service  ports.LeaderboardService
    clients  map[*websocket.Conn]bool
    mutex    sync.Mutex
    upgrader websocket.Upgrader
}

func NewWebSocketHub(service ports.LeaderboardService) *WebSocketHub {
    return &WebSocketHub{
        service: service,
        clients: make(map[*websocket.Conn]bool),
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true // Allow all origins for demo
            },
        },
    }
}

func (h *WebSocketHub) Run() {
    // Future implementation: Add channels for broadcasting messages
}

func (h *WebSocketHub) HandleConnection(w http.ResponseWriter, r *http.Request) {
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }
    defer conn.Close()

    h.mutex.Lock()
    h.clients[conn] = true
    h.mutex.Unlock()

    // Send initial leaderboard data
    rankings, err := h.service.GetRankings(r.Context())
    if err == nil {
        update := events.NewUpdate("full_update", nil, rankings)
        conn.WriteJSON(update)
    }

    // Handle incoming messages
    for {
        var player models.Player
        err := conn.ReadJSON(&player)
        if err != nil {
            log.Printf("Error reading message: %v", err)
            break
        }

        if err := h.service.UpdatePlayerScore(r.Context(), &player); err != nil {
            log.Printf("Error updating score: %v", err)
            continue
        }

        h.broadcastUpdate(&player)
    }

    h.mutex.Lock()
    delete(h.clients, conn)
    h.mutex.Unlock()
}

func (h *WebSocketHub) broadcastUpdate(player *models.Player) {
    rankings, err := h.service.GetRankings(context.Background())
    if err != nil {
        log.Printf("Error getting rankings: %v", err)
        return
    }

    update := events.NewUpdate("update", player, rankings)

    h.mutex.Lock()
    for conn := range h.clients {
        if err := conn.WriteJSON(update); err != nil {
            log.Printf("Error sending to client: %v", err)
            conn.Close()
            delete(h.clients, conn)
        }
    }
    h.mutex.Unlock()
}