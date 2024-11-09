package api

import (
	"net/http"
	"sales-analytics/internal/storage"

	"github.com/gorilla/websocket"
)

type Server struct {
	analytics *storage.AnalyticsStore
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
}

func NewServer(analytics *storage.AnalyticsStore) *Server {
	return &Server{
		analytics: analytics,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) BroadcastAnalytics() error {
	analytics, err := s.analytics.GetAnalytics()
	if err != nil {
		return err
	}

	for client := range s.clients {
		if err := client.WriteJSON(analytics); err != nil {
			delete(s.clients, client)
			client.Close()
		}
	}
	return nil
}