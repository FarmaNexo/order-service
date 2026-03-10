package postgres

import (
	"context"

	"github.com/farmanexo/order-service/internal/domain/entities"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderStatusHistoryRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewOrderStatusHistoryRepository(db *gorm.DB, logger *zap.Logger) *OrderStatusHistoryRepositoryImpl {
	return &OrderStatusHistoryRepositoryImpl{db: db, logger: logger}
}

func (r *OrderStatusHistoryRepositoryImpl) Create(ctx context.Context, history *entities.OrderStatusHistory) error {
	if history.ID == "" {
		history.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *OrderStatusHistoryRepositoryImpl) FindByOrderID(ctx context.Context, orderID string) ([]entities.OrderStatusHistory, error) {
	var history []entities.OrderStatusHistory
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at ASC").Find(&history).Error
	return history, err
}

var _ repositories.OrderStatusHistoryRepository = (*OrderStatusHistoryRepositoryImpl)(nil)
