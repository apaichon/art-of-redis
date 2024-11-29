// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var (
	rdb    *redis.Client
	ctx    = context.Background()
	secret = []byte("your-secret-key")
)

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/logout", logoutHandler).Methods("POST")
	r.HandleFunc("/check-auth", checkAuthHandler).Methods("GET")

	fmt.Println("Server is running on port 9000")
	http.ListenAndServe(":9000", r)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// In a real app, validate against DB
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	storedHash := hashedPassword // Simulate stored password

	if err := bcrypt.CompareHashAndPassword(storedHash, []byte(password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create session
	sessionID := uuid.New().String()
	userData, _ := json.Marshal(User{Username: username})

	// Store in Redis with 24 hour expiry
	rdb.Set(ctx, "session:"+sessionID, userData, 24*time.Hour)


	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("HX-Trigger", "authSuccess")
	w.Write([]byte(`<div id="authStatus">Logged in as ` + username + `</div>`))

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	fmt.Println("Cookie:", cookie)
	if err == nil {
		rdb.Del(ctx, "session:"+cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.Header().Set("HX-Trigger", "authLogout")
	w.Write([]byte(`<div id="authStatus">Logged out</div>`))
}

func checkAuthHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	val, err := rdb.Get(ctx, "session:"+cookie.Value).Result()
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var user User
	json.Unmarshal([]byte(val), &user)
	w.Write([]byte(`<div id="authStatus">Logged in as ` + user.Username + `</div>`))
}
