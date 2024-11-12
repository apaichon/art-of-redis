package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"luckydraw/internal/models"
	"luckydraw/internal/store"
	hub "luckydraw/internal/websocket"
	"luckydraw/pkg/utils"

	"github.com/gorilla/websocket"
)

type Handler struct {
	store    *store.RedisStore
	hub      *hub.Hub
	upgrader websocket.Upgrader
}

func NewHandler(store *store.RedisStore, hub *hub.Hub) *Handler {
	return &Handler{
		store: store,
		hub:   hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Handler) StartDraw(w http.ResponseWriter, r *http.Request) {
	draw := &models.Draw{
		ID:        utils.GenerateDrawID(),
		Number:    utils.GenerateNumber(),
		Status:    "spinning",
		CreatedAt: time.Now(),
	}
	fmt.Printf("draw: %+v\n", draw)

	if err := h.store.StoreDraw(draw); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast <- &models.Message{
		Type: "draw_update",
		Data: draw,
	}

	go h.processDraw(draw)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(draw)
}

func (h *Handler) processDraw(draw *models.Draw) {
	time.Sleep(5 * time.Second)

	draw.Status = "completed"
	draw.WinningNumber = utils.GenerateNumber()
	draw.CompletedAt = time.Now()

	fmt.Printf("completed draw: %+v\n", draw)

	h.store.StoreDraw(draw)

	h.hub.Broadcast <- &models.Message{
		Type: "draw_update",
		Data: draw,
	}
}

func (h *Handler) ClaimPrize(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var claimRequest struct {
		DrawID string `json:"draw_id"`
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&claimRequest); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate the claim
	draw, err := h.store.GetDraw(claimRequest.DrawID)
	if err != nil || draw.Status != "completed" {
		http.Error(w, "Draw not found or not completed", http.StatusNotFound)
		return
	}

	// Update the store with the claimed prize
	fmt.Printf("claimed draw: %+v\n", draw)

	if err := h.store.ClaimPrize(claimRequest.DrawID, claimRequest.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send a success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Prize claimed successfully"})
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Register the new connection with the hub
	h.hub.Register <- conn

	for {
		// Read messages from the WebSocket
		var msg models.Message
		if err := conn.ReadJSON(&msg); err != nil {
			break // Exit the loop on error
		}

		// Handle the message (you can add your own logic here)
		h.hub.Broadcast <- &msg
	}
}

func CORSHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin": "*",
	}
}
