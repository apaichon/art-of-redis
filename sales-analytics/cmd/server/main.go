package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"sales-analytics/internal/api"
	"sales-analytics/internal/storage"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize Redis store
	redisStore := storage.NewRedisStore("localhost:6379")
	defer redisStore.Close()

	// Initialize analytics store
	analyticsStore := storage.NewAnalyticsStore(redisStore)

	// Initialize server
	server := api.NewServer(analyticsStore)
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/ws", server.HandleWebSocket)
	router.HandleFunc("/api/sales", server.HandleSale).Methods("POST")
	router.HandleFunc("/api/sale-random", server.HandleSaleRandom).Methods("POST")
	router.HandleFunc("/api/remove-all", server.RemoveAll).Methods("POST")
	// Create server with router
	httpServer := &http.Server{
		Addr:    ":9003",
		Handler: router,
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		if err := httpServer.Close(); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()

	// Start server
	log.Println("Server starting on :9003")
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server stopped")
}
