package redis

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "ticket-reservation/internal/domain"
    redisClient "ticket-reservation/pkg/redis"
)

const reservationPrefix = "reservation:"

type ReservationRepository struct {
    client *redisClient.Client
    ttl    time.Duration
}

func NewReservationRepository(client *redisClient.Client, ttl time.Duration) *ReservationRepository {
    return &ReservationRepository{
        client: client,
        ttl:    ttl,
    }
}

func (r *ReservationRepository) Reserve(ctx context.Context, reservation *domain.Reservation) error {
    data, err := json.Marshal(reservation)
    if err != nil {
        return err
    }
    
    key := fmt.Sprintf("%s%s", reservationPrefix, reservation.TicketID)
    return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *ReservationRepository) GetByTicket(ctx context.Context, ticketID string) (*domain.Reservation, error) {
    data, err := r.client.Get(ctx, fmt.Sprintf("%s%s", reservationPrefix, ticketID)).Bytes()
    if err != nil {
        return nil, err
    }

    var reservation domain.Reservation
    if err := json.Unmarshal(data, &reservation); err != nil {
        return nil, err
    }
    return &reservation, nil
}