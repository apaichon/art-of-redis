# Session Management

## Overview
Session management is a crucial part of any web application. It is used to manage user sessions, authenticate users, and manage user sessions.

## Setup
### Backend
```bash
go mod tidy
```
Run the server
```bash
go run cmd/server.go
```
### Frontend
```bash
cd front
npm install
```
Run the frontend server
```bash
npm run dev
```

## Folder Structure

```
session-management/
├── auth/
│   ├── auth.go
│   ├── auth_test.go
├── cmd/
│   ├── server.go
```

## Class Diagram

```mermaid
classDiagram
    class Server {
        +redis *redis.Client
        +handleLogin(w http.ResponseWriter, r *http.Request)
        +handleLogout(w http.ResponseWriter, r *http.Request)
        +handleCheckAuth(w http.ResponseWriter, r *http.Request)
        +getSession(r *http.Request) *Session
        +generateSessionToken() string
    }

    class LoginRequest {
        +Email string
        +Password string
    }

    class Session {
        +UserID string
        +Email string
        +CreatedAt time.Time
    }

    class LoginHandler {
        +server *Server
        +HandleLogin(w http.ResponseWriter, r *http.Request)
    }

    class LogoutHandler {
        +server *Server
        +HandleLogout(w http.ResponseWriter, r *http.Request)
    }

    class CheckAuthHandler {
        +server *Server
        +HandleCheckAuth(w http.ResponseWriter, r *http.Request)
    }

    class ProtectedHandler {
        +HandleGet(w http.ResponseWriter, r *http.Request)
    }

    class main {
        +main()
    }

    Server --> LoginHandler
    Server --> LogoutHandler
    Server --> CheckAuthHandler
    Server --> ProtectedHandler
    LoginHandler --> LoginRequest
    Session --> LoginRequest
```


## 1. Login
### Api Flow
```mermaid
sequenceDiagram
    participant Client
    participant Server
    Client->>Server: POST /api/login
    Server->>Server: Validate login credentials
    Server->>Client: 200 OK
```
### Redis Flow
```mermaid
flowchart
    A[Client POST /api/login] --> B[Server validate login credentials]
    B --> C[Server generate session token]
    C --> D[Server store session token in Redis]
    D --> E[Server return 200 OK]
```
### Redis Commands
```bash
HSET session:<session_token> user_id <user_id> email <email> created_at <created_at>
```

### Go Code
```go
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
```

## 2. Logout
### Api Flow
```mermaid
sequenceDiagram
    participant Client
    participant Server
    Client->>Server: POST /api/logout
    Server->>Server: Delete session token from Redis
    Server->>Client: 200 OK
```
### Redis Flow
```mermaid
flowchart
    A[Client POST /api/logout] --> B[Server delete session token from Redis]
    B --> C[Server return 200 OK]
```
### Redis Commands
```bash
DEL session:<session_token>
```
### Go Code
```go
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.redis.Del(r.Context(), "session:"+s.getSession(r).SessionToken)
	w.WriteHeader(http.StatusOK)
}
```

## 3. Check Auth
### Api Flow
```mermaid
sequenceDiagram
    participant Client
    participant Server
    Client->>Server: POST /api/check-auth
    Server->>Server: Check session token in Redis
    Server->>Client: 200 OK
```
### Redis Flow
```mermaid
flowchart
    A[Client POST /api/check-auth] --> B[Server check session token in Redis]
    B --> C[Server return 200 OK]
```
### Redis Commands
```bash
HGETALL session:<session_token>
```
### Go Code
```go
func (s *Server) handleCheckAuth(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
``` 

## Summary
- Session management is a crucial part of any web application. It is used to manage user sessions, authenticate users, and manage user sessions.
- Redis is used to store session tokens.
- Go is used to implement the server.
- Axios is used to make requests to the server.
- Nodejs is used to run the frontend server.