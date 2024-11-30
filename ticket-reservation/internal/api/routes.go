package api

import (
    "fmt"
    "net/http"
    "ticket-reservation/internal/api/handlers"
    "ticket-reservation/internal/api/middleware"
    "ticket-reservation/internal/service"
    
    "github.com/gorilla/mux"
    "github.com/rs/cors"
)

type Router struct {
    ticketHandler      *handlers.TicketHandler
    waitingListHandler *handlers.WaitingListHandler
    notificationService *service.NotificationService
}

func NewRouter(
    ticketHandler *handlers.TicketHandler,
    waitingListHandler *handlers.WaitingListHandler,
) *Router {
    return &Router{
        ticketHandler:      ticketHandler,
        waitingListHandler: waitingListHandler,
    }
}

func (router *Router) Setup() http.Handler {
    r := mux.NewRouter()
    
    // Middleware
    r.Use(middleware.JSONContentType)
    
    // Tickets
    r.HandleFunc("/tickets", router.ticketHandler.List).Methods(http.MethodGet)
    r.HandleFunc("/tickets/reserve", router.ticketHandler.Reserve).Methods(http.MethodPost)
    
    // Waiting List
    r.HandleFunc("/waiting-list", router.waitingListHandler.Join).Methods(http.MethodPost)
    r.HandleFunc("/waiting-list/{ticketId}/{userId}", router.waitingListHandler.Leave).Methods(http.MethodDelete)
    
    // SSE endpoint for notifications
    r.HandleFunc("/notifications/{userId}", router.HandleSSE).Methods(http.MethodGet)
    
    // CORS
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Authorization"},
        AllowCredentials: true,
    })
    
    return c.Handler(r)
}

func (router *Router) HandleSSE(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    userID := vars["userId"]

    // Set headers for SSE
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    // Create channel for client
    events := make(chan []byte)
    defer close(events)
    
    // Add client to notification service
    router.notificationService.AddClient(userID, events)
    defer router.notificationService.RemoveClient(userID)
    
    // Stream events to client
    for {
        select {
        case event := <-events:
            fmt.Fprintf(w, "data: %s\n\n", event)
            w.(http.Flusher).Flush()
        case <-r.Context().Done():
            return
        }
    }
}