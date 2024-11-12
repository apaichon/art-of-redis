package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	redis *redis.Client
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Session struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func NewServer(redisAddr string) *Server {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr, // Redis server address
	})

	// Optionally, you can test the connection here
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err) // Handle error appropriately
	}

	return &Server{redis: rdb}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Hash the password for the mock user
	storedHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	mockUser := &Session{
		UserID:    "1",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword(storedHash, []byte(loginReq.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	fmt.Println("Login successful")
	// Generate session token
	sessionToken, err := s.generateSessionToken()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	fmt.Println("Session token:", sessionToken)

	// Create session
	sessionJSON, _ := json.Marshal(mockUser)
	err = s.redis.Set(r.Context(), "session:"+sessionToken, sessionJSON, 24*time.Hour).Err()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Respond with session token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_token": sessionToken,
		"message":       "Login successful",
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session token from request body instead of cookie
	var logoutReq struct {
		SessionToken string `json:"session_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&logoutReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Println("Logout request:", logoutReq.SessionToken)

	// Delete session from Redis
	err := s.redis.Del(r.Context(), "session:"+logoutReq.SessionToken).Err()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})
}

func (s *Server) handleCheckAuth(w http.ResponseWriter, r *http.Request) {
	// Get session token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract the token from the header (assuming Bearer token format)
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Retrieve the session using the token
	sessionData, err := s.redis.Get(r.Context(), "session:"+token).Result()
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var session Session
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Println("Session:", session)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (s *Server) generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

type LoginHandler struct {
	server *Server
}

func NewLoginHandler(server *Server) *LoginHandler {
	return &LoginHandler{server: server}
}

func (h *LoginHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	h.server.handleLogin(w, r)
}

type LogoutHandler struct {
	server *Server
}

func NewLogoutHandler(server *Server) *LogoutHandler {
	return &LogoutHandler{server: server}
}

func (h *LogoutHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	h.server.handleLogout(w, r)
}

type CheckAuthHandler struct {
	server *Server
}

func NewCheckAuthHandler(server *Server) *CheckAuthHandler {
	return &CheckAuthHandler{server: server}
}

func (h *CheckAuthHandler) HandleCheckAuth(w http.ResponseWriter, r *http.Request) {
	h.server.handleCheckAuth(w, r)
}

type ProtectedHandler struct{}

func NewProtectedHandler() *ProtectedHandler {
	return &ProtectedHandler{}
}

func (p *ProtectedHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"message": "This is a protected endpoint",
	})
}
