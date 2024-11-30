package service

import (
    "context"
    "time"
    "ticket-reservation/internal/domain"
    "ticket-reservation/internal/repository"
)

type WaitingListService struct {
    waitingListRepo repository.WaitingListRepository
}

func NewWaitingListService(waitingListRepo repository.WaitingListRepository) *WaitingListService {
    return &WaitingListService{waitingListRepo: waitingListRepo}
}

func (s *WaitingListService) AddToWaitingList(ctx context.Context, ticketID, userID string) error {
    item := &domain.WaitingListItem{
        TicketID: ticketID,
        UserID:   userID,
        AddedAt:  time.Now(),
    }
    return s.waitingListRepo.Add(ctx, item)
}

func (s *WaitingListService) NotifyNextInLine(ctx context.Context, ticketID string) error {
    items, err := s.waitingListRepo.GetByTicket(ctx, ticketID)
    if err != nil || len(items) == 0 {
        return err
    }

    // Implement notification logic here
    return nil
}

func (s *WaitingListService) RemoveFromWaitingList(ctx context.Context, ticketID, userID string) error {
    return s.waitingListRepo.Remove(ctx, ticketID, userID)
}