package repositories

import (
	"context"

	"github.com/farmanexo/order-service/internal/domain/entities"
)

type CartRepository interface {
	FindByUserID(ctx context.Context, userID string) ([]entities.CartItem, error)
	FindByID(ctx context.Context, id string) (*entities.CartItem, error)
	FindByUserAndProduct(ctx context.Context, userID, productID, pharmacyID string) (*entities.CartItem, error)
	Create(ctx context.Context, item *entities.CartItem) error
	Update(ctx context.Context, item *entities.CartItem) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
}
