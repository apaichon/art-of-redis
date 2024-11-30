package repository

import (
    "context"
    "ticket-reservation/internal/domain"
)

type TicketRepository interface {
    Create(ctx context.Context, ticket *domain.Ticket) error
    Get(ctx context.Context, id string) (*domain.Ticket, error)
    Update(ctx context.Context, ticket *domain.Ticket) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context) ([]*domain.Ticket, error)
}

type ReservationRepository interface {
    Reserve(ctx context.Context, reservation *domain.Reservation) error
    Cancel(ctx context.Context, ticketID, userID string) error
    GetByTicket(ctx context.Context, ticketID string) (*domain.Reservation, error)
    GetByUser(ctx context.Context, userID string) ([]*domain.Reservation, error)
}

type WaitingListRepository interface {
    Add(ctx context.Context, item *domain.WaitingListItem) error
    Remove(ctx context.Context, ticketID, userID string) error
    GetByTicket(ctx context.Context, ticketID string) ([]*domain.WaitingListItem, error)
    GetByUser(ctx context.Context, userID string) ([]*domain.WaitingListItem, error)
    GetPosition(ctx context.Context, ticketID, userID string) (int, error)
}