package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"luckydraw/internal/config"
	"luckydraw/internal/handlers"
	"luckydraw/internal/store"
	"luckydraw/internal/websocket"
)

func main() {
	cfg := config.New()
	
	store, err := store.NewRedisStore(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	
	hub := websocket.NewHub()
	go hub.Run()
	
	handler := handlers.NewHandler(store, hub)
	router := setupRoutes(handler)
	
	log.Printf("Server starting on %s", cfg.ServerAddr)
	log.Fatal(http.ListenAndServe(cfg.ServerAddr, router))
}

func setupRoutes(h *handlers.Handler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/draw/start", h.StartDraw).Methods("POST")
	router.HandleFunc("/api/draw/claim", h.ClaimPrize).Methods("POST")
	router.HandleFunc("/ws", h.HandleWebSocket)
	
	// Implement CORS origin
	headers := handlers.CORSHeaders()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, value := range headers {
				w.Header().Set(key, value)
			}
			next.ServeHTTP(w, r)
		})
	})
	
	return router
}