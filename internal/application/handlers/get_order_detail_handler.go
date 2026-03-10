package handlers

import (
	"context"

	"github.com/farmanexo/order-service/internal/application/queries"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/internal/shared/constants"
	"github.com/farmanexo/order-service/pkg/mediator"
	"go.uber.org/zap"
)

type GetOrderDetailHandler struct {
	orderRepo repositories.OrderRepository
	logger    *zap.Logger
}

func NewGetOrderDetailHandler(orderRepo repositories.OrderRepository, logger *zap.Logger) *GetOrderDetailHandler {
	return &GetOrderDetailHandler{orderRepo: orderRepo, logger: logger}
}

func (h *GetOrderDetailHandler) Handle(ctx context.Context, query queries.GetOrderDetailQuery) (*common.ApiResponse[responses.OrderDetailResponse], error) {
	order, err := h.orderRepo.FindByIDWithItems(ctx, query.OrderID)
	if err != nil || order == nil {
		return common.NotFoundResponse[responses.OrderDetailResponse]("Orden no encontrada"), nil
	}

	if order.UserID != query.UserID {
		return common.NotFoundResponse[responses.OrderDetailResponse]("Orden no encontrada"), nil
	}

	detail := buildOrderDetail(order)
	resp := common.OkResponse(detail)
	resp.AddMessageWithType(constants.CodeOrderRetrieved, constants.GetDescription(constants.CodeOrderRetrieved), constants.MessageTypeSuccess)
	return resp, nil
}

var _ mediator.RequestHandler[queries.GetOrderDetailQuery, responses.OrderDetailResponse] = (*GetOrderDetailHandler)(nil)
