package server

import (
    "encoding/json"
    "net/http"

	"github.com/gorilla/mux"
	"leaderboard/internal/ports"
)

type Handler struct {
    service ports.LeaderboardService
    hub     *WebSocketHub
}

func NewHandler(service ports.LeaderboardService) *Handler {
    hub := NewWebSocketHub(service)
    go hub.Run()
    
    return &Handler{
        service: service,
        hub:     hub,
    }
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
    r.HandleFunc("/api/leaderboard", h.handleGetLeaderboard).Methods("GET")
    r.HandleFunc("/ws", h.handleWebSocket)
    r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
}

func (h *Handler) handleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
    rankings, err := h.service.GetRankings(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(rankings)
}

func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    h.hub.HandleConnection(w, r)
}