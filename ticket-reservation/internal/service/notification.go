package service

import (
    "context"
    "encoding/json"
    "sync"
)

type NotificationService struct {
    clients map[string]chan []byte
    mu      sync.RWMutex
}

type Notification struct {
    Type    string      `json:"type"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func NewNotificationService() *NotificationService {
    return &NotificationService{
        clients: make(map[string]chan []byte),
    }
}

func (s *NotificationService) AddClient(userID string, ch chan []byte) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.clients[userID] = ch
}

func (s *NotificationService) RemoveClient(userID string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.clients, userID)
}

func (s *NotificationService) NotifyUser(userID string, notification *Notification) error {
    s.mu.RLock()
    ch, exists := s.clients[userID]
    s.mu.RUnlock()

    if !exists {
        return nil
    }

    data, err := json.Marshal(notification)
    if err != nil {
        return err
    }

    select {
    case ch <- data:
    default:
        // Channel is blocked, skip notification
    }
    return nil
}

func (s *NotificationService) NotifyReservationExpired(ctx context.Context, ticketID, userID string) {
    notification := &Notification{
        Type:    "reservation_expired",
        Message: "Your ticket reservation has expired",
        Data: map[string]string{
            "ticket_id": ticketID,
        },
    }
    s.NotifyUser(userID, notification)
}

func (s *NotificationService) NotifyWaitingListAvailable(ctx context.Context, ticketID string, waiters []string) {
    notification := &Notification{
        Type:    "ticket_available",
        Message: "A ticket you're waiting for is now available",
        Data: map[string]string{
            "ticket_id": ticketID,
        },
    }

    for _, userID := range waiters {
        s.NotifyUser(userID, notification)
    }
}