package repositories

import (
	"context"

	"github.com/farmanexo/order-service/internal/domain/entities"
)

type OrderStatusHistoryRepository interface {
	Create(ctx context.Context, history *entities.OrderStatusHistory) error
	FindByOrderID(ctx context.Context, orderID string) ([]entities.OrderStatusHistory, error)
}
