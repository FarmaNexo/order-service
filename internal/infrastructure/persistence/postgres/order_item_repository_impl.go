package postgres

import (
	"context"

	"github.com/farmanexo/order-service/internal/domain/entities"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderItemRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewOrderItemRepository(db *gorm.DB, logger *zap.Logger) *OrderItemRepositoryImpl {
	return &OrderItemRepositoryImpl{db: db, logger: logger}
}

func (r *OrderItemRepositoryImpl) CreateBatch(ctx context.Context, items []entities.OrderItem) error {
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *OrderItemRepositoryImpl) FindByOrderID(ctx context.Context, orderID string) ([]entities.OrderItem, error) {
	var items []entities.OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

var _ repositories.OrderItemRepository = (*OrderItemRepositoryImpl)(nil)
