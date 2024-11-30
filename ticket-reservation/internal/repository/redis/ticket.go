package redis

import (
    "context"
    "encoding/json"
    "ticket-reservation/internal/domain"
    redisClient "ticket-reservation/pkg/redis"
)

const ticketPrefix = "ticket:"

type TicketRepository struct {
    client *redisClient.Client
}

func NewTicketRepository(client *redisClient.Client) *TicketRepository {
    return &TicketRepository{client: client}
}

func (r *TicketRepository) Create(ctx context.Context, ticket *domain.Ticket) error {
    data, err := json.Marshal(ticket)
    if err != nil {
        return err
    }
    return r.client.Set(ctx, ticketPrefix+ticket.ID, data, 0).Err()
}

func (r *TicketRepository) Get(ctx context.Context, id string) (*domain.Ticket, error) {
    data, err := r.client.Get(ctx, ticketPrefix+id).Bytes()
    if err != nil {
        return nil, err
    }
    
    var ticket domain.Ticket
    if err := json.Unmarshal(data, &ticket); err != nil {
        return nil, err
    }
    return &ticket, nil
}

func (r *TicketRepository) Update(ctx context.Context, ticket *domain.Ticket) error {
    return r.Create(ctx, ticket)
}

func (r *TicketRepository) List(ctx context.Context) ([]*domain.Ticket, error) {
    keys, err := r.client.Keys(ctx, ticketPrefix+"*").Result()
    if err != nil {
        return nil, err
    }

    tickets := make([]*domain.Ticket, 0, len(keys))
    for _, key := range keys {
        ticket, err := r.Get(ctx, key[len(ticketPrefix):])
        if err != nil {
            continue
        }
        tickets = append(tickets, ticket)
    }
    return tickets, nil
}