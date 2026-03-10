package postgres

import (
	"context"

	"github.com/farmanexo/order-service/internal/domain/entities"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CartRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewCartRepository(db *gorm.DB, logger *zap.Logger) *CartRepositoryImpl {
	return &CartRepositoryImpl{db: db, logger: logger}
}

func (r *CartRepositoryImpl) FindByUserID(ctx context.Context, userID string) ([]entities.CartItem, error) {
	var items []entities.CartItem
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at ASC").Find(&items).Error
	return items, err
}

func (r *CartRepositoryImpl) FindByID(ctx context.Context, id string) (*entities.CartItem, error) {
	var item entities.CartItem
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartRepositoryImpl) FindByUserAndProduct(ctx context.Context, userID, productID, pharmacyID string) (*entities.CartItem, error) {
	var item entities.CartItem
	err := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ? AND pharmacy_id = ?", userID, productID, pharmacyID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartRepositoryImpl) Create(ctx context.Context, item *entities.CartItem) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *CartRepositoryImpl) Update(ctx context.Context, item *entities.CartItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *CartRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entities.CartItem{}).Error
}

func (r *CartRepositoryImpl) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entities.CartItem{}).Error
}

var _ repositories.CartRepository = (*CartRepositoryImpl)(nil)
