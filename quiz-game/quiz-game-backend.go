// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	redis     *redis.Client
	ctx       context.Context
	upgrader  websocket.Upgrader
	games     map[string]*Game
	gamesMux  sync.RWMutex
}

type Game struct {
	ID          string             `json:"id"`
	Pin         string             `json:"pin"`
	Host        *websocket.Conn    `json:"-"`
	Players     map[string]*Player `json:"players"`
	Questions   []Question         `json:"questions"`
	CurrentQ    int               `json:"currentQuestion"`
	Status      string            `json:"status"` // waiting, active, finished
	StartTime   time.Time         `json:"startTime"`
	PlayersMux  sync.RWMutex
}

type Player struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Score    int             `json:"score"`
	Conn     *websocket.Conn `json:"-"`
	Answers  []PlayerAnswer  `json:"answers"`
}

type Question struct {
	ID         string   `json:"id"`
	Text       string   `json:"text"`
	Options    []string `json:"options"`
	Correct    int      `json:"correct"`
	TimeLimit  int      `json:"timeLimit"`
	Points     int      `json:"points"`
}

type PlayerAnswer struct {
	QuestionID string        `json:"questionId"`
	Answer     int           `json:"answer"`
	Time       time.Duration `json:"time"`
	Points     int          `json:"points"`
}

func NewServer() *Server {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	return &Server{
		redis: rdb,
		ctx:   context.Background(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		games: make(map[string]*Game),
	}
}

// CreateGame creates a new quiz game
func (s *Server) handleCreateGame(w http.ResponseWriter, r *http.Request) {
	var questions []Question
	if err := json.NewDecoder(r.Body).Decode(&questions); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	game := &Game{
		ID:       fmt.Sprintf("game:%d", time.Now().UnixNano()),
		Pin:      fmt.Sprintf("%06d", time.Now().UnixNano()%1000000),
		Players:  make(map[string]*Player),
		Status:   "waiting",
		Questions: questions,
	}

	s.gamesMux.Lock()
	s.games[game.ID] = game
	s.gamesMux.Unlock()

	// Store game in Redis
	gameJSON, _ := json.Marshal(game)
	s.redis.Set(s.ctx, game.ID, gameJSON, 24*time.Hour)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"gameId": game.ID,
		"pin":    game.Pin,
	})
}

// Host WebSocket connection
func (s *Server) handleHostConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	gameID := r.URL.Query().Get("gameId")
	s.gamesMux.RLock()
	game, exists := s.games[gameID]
	s.gamesMux.RUnlock()

	if !exists {
		conn.WriteJSON(map[string]string{"error": "Game not found"})
		return
	}

	game.Host = conn

	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg["type"].(string) {
		case "start_game":
			s.startGame(game)
		case "next_question":
			s.nextQuestion(game)
		}
	}
}

// Player WebSocket connection
func (s *Server) handlePlayerConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Wait for player to join with pin
	var joinMsg map[string]string
	if err := conn.ReadJSON(&joinMsg); err != nil {
		return
	}

	pin := joinMsg["pin"]
	playerName := joinMsg["name"]

	// Find game by pin
	var game *Game
	s.gamesMux.RLock()
	for _, g := range s.games {
		if g.Pin == pin {
			game = g
			break
		}
	}
	s.gamesMux.RUnlock()

	if game == nil {
		conn.WriteJSON(map[string]string{"error": "Game not found"})
		return
	}

	// Create player
	player := &Player{
		ID:      fmt.Sprintf("player:%d", time.Now().UnixNano()),
		Name:    playerName,
		Conn:    conn,
		Answers: make([]PlayerAnswer, 0),
	}

	// Add player to game
	game.PlayersMux.Lock()
	game.Players[player.ID] = player
	game.PlayersMux.Unlock()

	// Notify host of new player
	if game.Host != nil {
		game.Host.WriteJSON(map[string]interface{}{
			"type":   "player_joined",
			"player": map[string]string{
				"id":   player.ID,
				"name": player.Name,
			},
		})
	}

	// Handle player messages
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg["type"].(string) {
		case "answer":
			s.handlePlayerAnswer(game, player, msg)
		}
	}

	// Remove player on disconnect
	game.PlayersMux.Lock()
	delete(game.Players, player.ID)
	game.PlayersMux.Unlock()
}

func (s *Server) startGame(game *Game) {
	game.Status = "active"
	game.StartTime = time.Now()
	game.CurrentQ = 0

	// Notify all players
	game.PlayersMux.RLock()
	for _, player := range game.Players {
		player.Conn.WriteJSON(map[string]interface{}{
			"type": "game_started",
		})
	}
	game.PlayersMux.RUnlock()

	// Send first question
	s.sendQuestion(game)
}

func (s *Server) nextQuestion(game *Game) {
	if game.CurrentQ >= len(game.Questions)-1 {
		s.endGame(game)
		return
	}

	game.CurrentQ++
	s.sendQuestion(game)
}

func (s *Server) sendQuestion(game *Game) {
	question := game.Questions[game.CurrentQ]
	questionMsg := map[string]interface{}{
		"type":      "question",
		"text":      question.Text,
		"options":   question.Options,
		"timeLimit": question.TimeLimit,
	}

	// Send to all players
	game.PlayersMux.RLock()
	for _, player := range game.Players {
		player.Conn.WriteJSON(questionMsg)
	}
	game.PlayersMux.RUnlock()

	// Start timer for question
	go func() {
		time.Sleep(time.Duration(question.TimeLimit) * time.Second)
		s.sendQuestionResults(game)
	}()
}

func (s *Server) handlePlayerAnswer(game *Game, player *Player, msg map[string]interface{}) {
	question := game.Questions[game.CurrentQ]
	answer := int(msg["answer"].(float64))
	answerTime := time.Duration(msg["time"].(float64))

	// Calculate points based on speed and correctness
	points := 0
	if answer == question.Correct {
		timeBonus := float64(question.TimeLimit*1000-int(answerTime)) / float64(question.TimeLimit*1000)
		points = int(float64(question.Points) * timeBonus)
	}

	playerAnswer := PlayerAnswer{
		QuestionID: question.ID,
		Answer:     answer,
		Time:      answerTime,
		Points:    points,
	}

	player.Answers = append(player.Answers, playerAnswer)
	player.Score += points

	// Store answer in Redis
	answerKey := fmt.Sprintf("game:%s:question:%s:player:%s", game.ID, question.ID, player.ID)
	answerJSON, _ := json.Marshal(playerAnswer)
	s.redis.Set(s.ctx, answerKey, answerJSON, 24*time.Hour)
}

func (s *Server) sendQuestionResults(game *Game) {
	question := game.Questions[game.CurrentQ]
	results := make(map[string]interface{})
	results["type"] = "question_results"
	results["correct"] = question.Correct

	playerResults := make([]map[string]interface{}, 0)
	game.PlayersMux.RLock()
	for _, player := range game.Players {
		if len(player.Answers) > game.CurrentQ {
			answer := player.Answers[game.CurrentQ]
			playerResults = append(playerResults, map[string]interface{}{
				"playerId": player.ID,
				"name":     player.Name,
				"answer":   answer.Answer,
				"points":   answer.Points,
				"score":    player.Score,
			})
		}
	}
	game.PlayersMux.RUnlock()

	results["players"] = playerResults

	// Send results to all
	if game.Host != nil {
		game.Host.WriteJSON(results)
	}
	game.PlayersMux.RLock()
	for _, player := range game.Players {
		player.Conn.WriteJSON(results)
	}
	game.PlayersMux.RUnlock()
}

func (s *Server) endGame(game *Game) {
	game.Status = "finished"

	// Calculate final scores and rankings
	type PlayerRank struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Score int    `json:"score"`
		Rank  int    `json:"rank"`
	}

	rankings := make([]PlayerRank, 0)
	game.PlayersMux.RLock()
	for _, player := range game.Players {
		rankings = append(rankings, PlayerRank{
			ID:    player.ID,
			Name:  player.Name,
			Score: player.Score,
		})
	}
	game.PlayersMux.RUnlock()

	// Sort rankings
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].Score > rankings[j].Score
	})

	// Assign ranks
	for i := range rankings {
		rankings[i].Rank = i + 1
	}

	// Send final results
	finalResults := map[string]interface{}{
		"type":     "game_over",
		"rankings": rankings,
	}

	if game.Host != nil {
		game.Host.WriteJSON(finalResults)
	}
	game.PlayersMux.RLock()
	for _, player := range game.Players {
		player.Conn.WriteJSON(finalResults)
	}
	game.PlayersMux.RUnlock()

	// Store results in Redis
	resultsJSON, _ := json.Marshal(finalResults)
	s.redis.Set(s.ctx, fmt.Sprintf("game:%s:results", game.ID), resultsJSON, 24*time.Hour)
}

func main() {
	server := NewServer()
	router := mux.NewRouter()

	router.HandleFunc("/api/games", server.handleCreateGame).Methods("POST")
	router.HandleFunc("/ws/host", server.handleHostConnection)
	router.HandleFunc("/ws/player", server.handlePlayerConnection)

	log.Fatal(http.ListenAndServe(":8080", router))
}
