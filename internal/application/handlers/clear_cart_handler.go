package handlers

import (
	"context"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/internal/shared/constants"
	"github.com/farmanexo/order-service/pkg/mediator"
	"go.uber.org/zap"
)

type ClearCartHandler struct {
	cartRepo repositories.CartRepository
	logger   *zap.Logger
}

func NewClearCartHandler(cartRepo repositories.CartRepository, logger *zap.Logger) *ClearCartHandler {
	return &ClearCartHandler{cartRepo: cartRepo, logger: logger}
}

func (h *ClearCartHandler) Handle(ctx context.Context, cmd commands.ClearCartCommand) (*common.ApiResponse[responses.EmptyResponse], error) {
	if err := h.cartRepo.DeleteByUserID(ctx, cmd.UserID); err != nil {
		h.logger.Error("Error vaciando carrito", zap.Error(err))
		return common.InternalServerErrorResponse[responses.EmptyResponse]("Error vaciando carrito"), nil
	}

	resp := common.OkResponse(responses.EmptyResponse{})
	resp.AddMessageWithType(constants.CodeCartCleared, constants.GetDescription(constants.CodeCartCleared), constants.MessageTypeSuccess)
	return resp, nil
}

var _ mediator.RequestHandler[commands.ClearCartCommand, responses.EmptyResponse] = (*ClearCartHandler)(nil)
