package repositories

import (
	"context"

	"github.com/farmanexo/order-service/internal/domain/entities"
)

type OrderRepository interface {
	Create(ctx context.Context, order *entities.Order) error
	FindByID(ctx context.Context, id string) (*entities.Order, error)
	FindByIDWithItems(ctx context.Context, id string) (*entities.Order, error)
	FindByUserID(ctx context.Context, userID string, status string, page, limit int) ([]entities.Order, int64, error)
	FindByPharmacyID(ctx context.Context, pharmacyID string, status string, page, limit int) ([]entities.Order, int64, error)
	FindAll(ctx context.Context, userID, pharmacyID, status, dateFrom, dateTo string, page, limit int) ([]entities.Order, int64, error)
	Update(ctx context.Context, order *entities.Order) error
	GetNextOrderNumber(ctx context.Context) (int64, error)
	GetOrderStats(ctx context.Context, dateFrom, dateTo string) (*OrderStats, error)
}

type OrderStats struct {
	TotalOrders  int64                `json:"total_orders"`
	TotalRevenue float64              `json:"total_revenue"`
	ByStatus     map[string]int64     `json:"by_status"`
	TopPharmacies []TopPharmacyStat   `json:"top_pharmacies"`
	AvgOrderValue float64             `json:"avg_order_value"`
}

type TopPharmacyStat struct {
	PharmacyID   string  `json:"pharmacy_id"`
	PharmacyName string  `json:"pharmacy_name"`
	OrdersCount  int64   `json:"orders_count"`
	Revenue      float64 `json:"revenue"`
}
