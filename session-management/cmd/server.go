package main

import (
	"log"
	"net/http"
	"session-management/auth" // Adjust the import path as necessary
)

func main() {
	// Initialize the Redis server connection
	redisAddr := "127.0.0.1:6379" // Change this to your Redis server address if needed
	server := auth.NewServer(redisAddr)

	// Set up your HTTP handlers

	loginHandler := auth.NewLoginHandler(server)
	logoutHandler := auth.NewLogoutHandler(server)
	checkAuthHandler := auth.NewCheckAuthHandler(server)
	protectedHandler := auth.NewProtectedHandler()

	// Serve static files from the 'public' directory

	http.HandleFunc("/api/login", loginHandler.HandleLogin)
	http.HandleFunc("/api/logout", logoutHandler.HandleLogout)
	http.HandleFunc("/api/check-auth", checkAuthHandler.HandleCheckAuth)
	http.HandleFunc("/api/protected", protectedHandler.HandleGet)

	// Start the HTTP server
	log.Println("Starting server on :9001")
	if err := http.ListenAndServe(":9001", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
