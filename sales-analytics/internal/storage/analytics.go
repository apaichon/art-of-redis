package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"sales-analytics/internal/models"
)

type AnalyticsStore struct {
	redis *RedisStore
}

func NewAnalyticsStore(redis *RedisStore) *AnalyticsStore {
	return &AnalyticsStore{redis: redis}
}

func (as *AnalyticsStore) RecordTicketSale(sale models.TicketSale) error {
	saleJSON, err := json.Marshal(sale)
	if err != nil {
		return fmt.Errorf("failed to marshal sale: %w", err)
	}

	pipe := as.redis.Pipeline()

	// Store complete sale object in a list
	pipe.LPush(as.redis.ctx, "recent_sales", saleJSON)
	pipe.LTrim(as.redis.ctx, "recent_sales", 0, 99)

	// Update aggregated metrics
	pipe.Incr(as.redis.ctx, "total_tickets_sold")
	pipe.IncrByFloat(as.redis.ctx, "total_revenue", sale.Price)

	// Update category metrics
	pipe.Incr(as.redis.ctx, fmt.Sprintf("category:%s:count", sale.Category))
	pipe.IncrByFloat(as.redis.ctx, fmt.Sprintf("category:%s:revenue", sale.Category), sale.Price)

	// Update hourly metrics
	hour := sale.Timestamp.Format("2006-01-02:15")
	pipe.Incr(as.redis.ctx, fmt.Sprintf("hourly:%s:count", hour))
	pipe.IncrByFloat(as.redis.ctx, fmt.Sprintf("hourly:%s:revenue", hour), sale.Price)

	// Update concert metrics
	pipe.Incr(as.redis.ctx, fmt.Sprintf("concert:%s:count", sale.ConcertID))
	pipe.IncrByFloat(as.redis.ctx, fmt.Sprintf("concert:%s:revenue", sale.ConcertID), sale.Price)
	pipe.HSet(as.redis.ctx, "concert_names", sale.ConcertID, sale.ConcertName)

	_, err = pipe.Exec(as.redis.ctx)
	return err
}

func (as *AnalyticsStore) GetAnalytics() (*models.SalesAnalytics, error) {
	analytics := &models.SalesAnalytics{
		SalesByCategory:  make(map[string]int),
		CategoryRevenue:  make(map[string]float64),
		HourlySalesCount: make(map[string]int),
	}

	// Get total metrics
	totalRevenue, err := as.redis.Get("total_revenue").Float64()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}
	analytics.TotalRevenue = totalRevenue

	ticketsSold, err := as.redis.Get("total_tickets_sold").Int()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get tickets sold: %w", err)
	}
	analytics.TicketsSold = ticketsSold

	// Get recent sales
	if err := as.getRecentSales(analytics); err != nil {
		return nil, err
	}

	// Get category stats
	if err := as.getCategoryStats(analytics); err != nil {
		return nil, err
	}

	// Get hourly revenue
	if err := as.getHourlyRevenue(analytics); err != nil {
		return nil, err
	}

	// Get top concerts
	if err := as.getTopConcerts(analytics); err != nil {
		return nil, err
	}

	return analytics, nil
}

func (as *AnalyticsStore) getRecentSales(analytics *models.SalesAnalytics) error {
	recentSalesJSON, err := as.redis.LRange("recent_sales", 0, 9).Result()
	if err != nil {
		return fmt.Errorf("failed to get recent sales: %w", err)
	}

	for _, saleJSON := range recentSalesJSON {
		var sale models.TicketSale
		if err := json.Unmarshal([]byte(saleJSON), &sale); err != nil {
			return fmt.Errorf("failed to unmarshal sale: %w", err)
		}
		analytics.RecentSales = append(analytics.RecentSales, sale)
	}
	return nil
}

func (as *AnalyticsStore) getCategoryStats(analytics *models.SalesAnalytics) error {
	categories := []string{"VIP", "Standard", "Economy"}
	for _, category := range categories {
		count, err := as.redis.Get(fmt.Sprintf("category:%s:count", category)).Int()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get category count: %w", err)
		}

		revenue, err := as.redis.Get(fmt.Sprintf("category:%s:revenue", category)).Float64()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get category revenue: %w", err)
		}

		analytics.SalesByCategory[category] = count
		analytics.CategoryRevenue[category] = revenue
	}
	return nil
}

func (as *AnalyticsStore) getHourlyRevenue(analytics *models.SalesAnalytics) error {
	now := time.Now()
	for i := 23; i >= 0; i-- {
		hour := now.Add(time.Duration(-i) * time.Hour).Format("2006-01-02:15")
		revenue, err := as.redis.Get(fmt.Sprintf("hourly:%s:revenue", hour)).Float64()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get hourly revenue: %w", err)
		}

		count, err := as.redis.Get(fmt.Sprintf("hourly:%s:count", hour)).Int()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get hourly count: %w", err)
		}

		analytics.RevenueByHour = append(analytics.RevenueByHour, models.HourlyRevenue{
			Hour:      hour[11:],
			Revenue:   revenue,
			SaleCount: count,
		})
	}
	return nil
}

func (as *AnalyticsStore) getTopConcerts(analytics *models.SalesAnalytics) error {
	concertNames, err := as.redis.HGetAll("concert_names").Result()
	if err != nil {
		return fmt.Errorf("failed to get concert names: %w", err)
	}

	for id, name := range concertNames {
		revenue, err := as.redis.Get(fmt.Sprintf("concert:%s:revenue", id)).Float64()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get concert revenue: %w", err)
		}

		count, err := as.redis.Get(fmt.Sprintf("concert:%s:count", id)).Int()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get concert count: %w", err)
		}

		analytics.TopConcerts = append(analytics.TopConcerts, models.ConcertSales{
			ConcertID:   id,
			ConcertName: name,
			Revenue:     revenue,
			TicketsSold: count,
		})
	}
	return nil
}


func (as *AnalyticsStore) RemoveAll() error {
	// Flush all keys in the current Redis database
	return as.redis.RemoveAll().Err()
}
