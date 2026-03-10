package repositories

import (
	"context"

	"github.com/farmanexo/order-service/internal/domain/entities"
)

type OrderItemRepository interface {
	CreateBatch(ctx context.Context, items []entities.OrderItem) error
	FindByOrderID(ctx context.Context, orderID string) ([]entities.OrderItem, error)
}
