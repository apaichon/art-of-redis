package service

import (
	"context"
	"errors"
	"ticket-reservation/internal/domain"
	"ticket-reservation/internal/repository"
	"time"
)

var ErrTicketAlreadyReserved = errors.New("ticket already reserved")

type TicketService struct {
	ticketRepo      repository.TicketRepository
	reservationRepo repository.ReservationRepository
	waitingListRepo repository.WaitingListRepository
}

func NewTicketService(
	ticketRepo repository.TicketRepository,
	reservationRepo repository.ReservationRepository,
	waitingListRepo repository.WaitingListRepository,
) *TicketService {
	return &TicketService{
		ticketRepo:      ticketRepo,
		reservationRepo: reservationRepo,
		waitingListRepo: waitingListRepo,
	}
}

func (s *TicketService) ReserveTicket(ctx context.Context, req *domain.ReservationRequest) error {
	ticket, err := s.ticketRepo.Get(ctx, req.TicketID)
	if err != nil {
		return err
	}

	if ticket.Reserved {
		return ErrTicketAlreadyReserved
	}

	ticket.Reserved = true
	ticket.ReservedBy = req.UserID
	ticket.ReservedAt = time.Now()
	ticket.ExpiresAt = ticket.ReservedAt.Add(30 * time.Minute)

	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return err
	}

	reservation := &domain.Reservation{
		TicketID:  req.TicketID,
		UserID:    req.UserID,
		CreatedAt: ticket.ReservedAt,
		ExpiresAt: ticket.ExpiresAt,
	}

	return s.reservationRepo.Reserve(ctx, reservation)
}

func (s *TicketService) List(ctx context.Context) ([]*domain.Ticket, error) {
	return s.ticketRepo.List(ctx)
}
