package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "ticket-reservation/config"
    "ticket-reservation/internal/api"
    "ticket-reservation/internal/api/handlers"
    "ticket-reservation/internal/repository/redis"
    "ticket-reservation/internal/service"
    redisClient "ticket-reservation/pkg/redis"
)

func main() {
    // Load config
    cfg := config.LoadConfig()
    
    // Setup Redis client
    client, err := redisClient.NewClient(cfg)
    if err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }
    defer client.Close()
    
    // Setup repositories
    ticketRepo := redis.NewTicketRepository(client)
    reservationRepo := redis.NewReservationRepository(client, cfg.ReservationTTL)
    waitingListRepo := redis.NewWaitingListRepository(client)
    
    // Setup services
    ticketService := service.NewTicketService(ticketRepo, reservationRepo, waitingListRepo)
    waitingListService := service.NewWaitingListService(waitingListRepo)
    
    // Setup handlers
    ticketHandler := handlers.NewTicketHandler(ticketService)
    waitingListHandler := handlers.NewWaitingListHandler(waitingListService)
    
    // Setup router
    router := api.NewRouter(ticketHandler, waitingListHandler)
    
    // Create server
    srv := &http.Server{
        Addr:    ":" + cfg.ServerPort,
        Handler: router.Setup(),
    }
    
    // Start server
    go func() {
        log.Printf("Server starting on port %s", cfg.ServerPort)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Failed to start server:", err)
        }
    }()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server stopped gracefully")
}