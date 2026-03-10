package handlers

import (
	"context"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/application/queries"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/pkg/mediator"
	"go.uber.org/zap"
)

type RemoveCartItemHandler struct {
	cartRepo repositories.CartRepository
	logger   *zap.Logger
}

func NewRemoveCartItemHandler(cartRepo repositories.CartRepository, logger *zap.Logger) *RemoveCartItemHandler {
	return &RemoveCartItemHandler{cartRepo: cartRepo, logger: logger}
}

func (h *RemoveCartItemHandler) Handle(ctx context.Context, cmd commands.RemoveCartItemCommand) (*common.ApiResponse[responses.CartResponse], error) {
	item, err := h.cartRepo.FindByID(ctx, cmd.ItemID)
	if err != nil || item == nil {
		return common.NotFoundResponse[responses.CartResponse]("Item no encontrado en el carrito"), nil
	}

	if item.UserID != cmd.UserID {
		return common.NotFoundResponse[responses.CartResponse]("Item no encontrado en el carrito"), nil
	}

	if err := h.cartRepo.Delete(ctx, cmd.ItemID); err != nil {
		h.logger.Error("Error eliminando item del carrito", zap.Error(err))
		return common.InternalServerErrorResponse[responses.CartResponse]("Error eliminando item"), nil
	}

	getCartHandler := NewGetCartHandler(h.cartRepo, h.logger)
	return getCartHandler.Handle(ctx, queries.GetCartQuery{UserID: cmd.UserID})
}

var _ mediator.RequestHandler[commands.RemoveCartItemCommand, responses.CartResponse] = (*RemoveCartItemHandler)(nil)
