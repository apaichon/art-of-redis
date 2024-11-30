package redis

import (
    "context"
    "encoding/json"
    "fmt"
    "ticket-reservation/internal/domain"
    redisClient "ticket-reservation/pkg/redis"
)

const waitingListPrefix = "waiting:"

type WaitingListRepository struct {
    client *redisClient.Client
}

func NewWaitingListRepository(client *redisClient.Client) *WaitingListRepository {
    return &WaitingListRepository{client: client}
}

func (r *WaitingListRepository) Add(ctx context.Context, item *domain.WaitingListItem) error {
    data, err := json.Marshal(item)
    if err != nil {
        return err
    }
    
    key := fmt.Sprintf("%s%s", waitingListPrefix, item.TicketID)
    return r.client.RPush(ctx, key, data).Err()
}

func (r *WaitingListRepository) GetByTicket(ctx context.Context, ticketID string) ([]*domain.WaitingListItem, error) {
    key := fmt.Sprintf("%s%s", waitingListPrefix, ticketID)
    data, err := r.client.LRange(ctx, key, 0, -1).Result()
    if err != nil {
        return nil, err
    }

    items := make([]*domain.WaitingListItem, 0, len(data))
    for i, itemData := range data {
        var item domain.WaitingListItem
        if err := json.Unmarshal([]byte(itemData), &item); err != nil {
            continue
        }
        item.Position = i + 1
        items = append(items, &item)
    }
    return items, nil
}

func (r *WaitingListRepository) Remove(ctx context.Context, ticketID, userID string) error {
    key := fmt.Sprintf("%s%s", waitingListPrefix, ticketID)
    items, err := r.GetByTicket(ctx, ticketID)
    if err != nil {
        return err
    }

    for i, item := range items {
        if item.UserID == userID {
            return r.client.LRem(ctx, key, 1, items[i]).Err()
        }
    }
    return nil
}