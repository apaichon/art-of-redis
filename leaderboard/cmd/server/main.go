package main

import (
	"context"
	"log"
	"net/http"

	"leaderboard/internal/config"
	"leaderboard/internal/repository"
	"leaderboard/internal/server"
	"leaderboard/internal/service"

	"github.com/gorilla/mux"
)

func main() {

	cfg := config.New()

	repo, err := repository.NewRedisRepository(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	// repo.ClearData(ctx)

	leaderboardService := service.NewLeaderboardService(repo)
	leaderboardService.RemoveLeaderboard(ctx)
	handler := server.NewHandler(leaderboardService)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	log.Printf("Server starting on %s", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}
