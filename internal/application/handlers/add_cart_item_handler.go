package handlers

import (
	"context"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/application/queries"
	"github.com/farmanexo/order-service/internal/domain/entities"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/farmanexo/order-service/internal/domain/services"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/internal/shared/constants"
	"github.com/farmanexo/order-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AddCartItemHandler struct {
	cartRepo       repositories.CartRepository
	pharmacyClient services.PharmacyClient
	logger         *zap.Logger
}

func NewAddCartItemHandler(
	cartRepo repositories.CartRepository,
	pharmacyClient services.PharmacyClient,
	logger *zap.Logger,
) *AddCartItemHandler {
	return &AddCartItemHandler{cartRepo: cartRepo, pharmacyClient: pharmacyClient, logger: logger}
}

func (h *AddCartItemHandler) Handle(ctx context.Context, cmd commands.AddCartItemCommand) (*common.ApiResponse[responses.CartResponse], error) {
	// Verify product exists in pharmacy inventory
	inventoryItem, err := h.pharmacyClient.GetInventoryItem(ctx, cmd.PharmacyID, cmd.ProductID)
	if err != nil || inventoryItem == nil {
		h.logger.Warn("Producto no encontrado en inventario de farmacia",
			zap.String("product_id", cmd.ProductID),
			zap.String("pharmacy_id", cmd.PharmacyID),
		)
		return common.BadRequestResponse[responses.CartResponse](constants.CodeProductNotFound, "Producto no disponible en esta farmacia"), nil
	}

	if inventoryItem.Stock < cmd.Quantity {
		return common.BadRequestResponse[responses.CartResponse](constants.CodeInsufficientStock, "Stock insuficiente. Disponible: "+string(rune(inventoryItem.Stock+'0'))), nil
	}

	// Check if item already exists in cart
	existing, _ := h.cartRepo.FindByUserAndProduct(ctx, cmd.UserID, cmd.ProductID, cmd.PharmacyID)
	if existing != nil {
		existing.Quantity += cmd.Quantity
		existing.UnitPrice = inventoryItem.Price
		if err := h.cartRepo.Update(ctx, existing); err != nil {
			h.logger.Error("Error actualizando item del carrito", zap.Error(err))
			return common.InternalServerErrorResponse[responses.CartResponse]("Error actualizando carrito"), nil
		}
	} else {
		newItem := &entities.CartItem{
			ID:           uuid.New().String(),
			UserID:       cmd.UserID,
			ProductID:    cmd.ProductID,
			PharmacyID:   cmd.PharmacyID,
			Quantity:     cmd.Quantity,
			UnitPrice:    inventoryItem.Price,
			ProductName:  inventoryItem.ProductName,
			PharmacyName: inventoryItem.PharmacyName,
		}
		if err := h.cartRepo.Create(ctx, newItem); err != nil {
			h.logger.Error("Error creando item del carrito", zap.Error(err))
			return common.InternalServerErrorResponse[responses.CartResponse]("Error agregando al carrito"), nil
		}
	}

	// Return updated cart using GetCartHandler logic (inline)
	return h.getCartResponse(ctx, cmd.UserID)
}

func (h *AddCartItemHandler) getCartResponse(ctx context.Context, userID string) (*common.ApiResponse[responses.CartResponse], error) {
	getCartHandler := NewGetCartHandler(h.cartRepo, h.logger)
	return getCartHandler.Handle(ctx, queries.GetCartQuery{UserID: userID})
}

var _ mediator.RequestHandler[commands.AddCartItemCommand, responses.CartResponse] = (*AddCartItemHandler)(nil)
