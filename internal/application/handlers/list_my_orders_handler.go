package handlers

import (
	"context"
	"math"

	"github.com/farmanexo/order-service/internal/application/queries"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/internal/shared/constants"
	"github.com/farmanexo/order-service/pkg/mediator"
	"go.uber.org/zap"
)

type ListMyOrdersHandler struct {
	orderRepo repositories.OrderRepository
	logger    *zap.Logger
}

func NewListMyOrdersHandler(orderRepo repositories.OrderRepository, logger *zap.Logger) *ListMyOrdersHandler {
	return &ListMyOrdersHandler{orderRepo: orderRepo, logger: logger}
}

func (h *ListMyOrdersHandler) Handle(ctx context.Context, query queries.ListMyOrdersQuery) (*common.ApiResponse[responses.OrderListResponse], error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	limit := query.Limit
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	orders, total, err := h.orderRepo.FindByUserID(ctx, query.UserID, query.Status, page, limit)
	if err != nil {
		h.logger.Error("Error listando órdenes", zap.Error(err))
		return common.InternalServerErrorResponse[responses.OrderListResponse]("Error listando órdenes"), nil
	}

	summaries := make([]responses.OrderSummaryResponse, 0, len(orders))
	for _, order := range orders {
		summaries = append(summaries, responses.OrderSummaryResponse{
			OrderID:     order.ID,
			OrderNumber: order.OrderNumber,
			PharmacyID:  order.PharmacyID,
			Total:       order.Total,
			Status:      order.Status,
			ItemsCount:  len(order.Items),
			CreatedAt:   order.CreatedAt,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	listResp := responses.OrderListResponse{
		Orders:     summaries,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	resp := common.OkResponse(listResp)
	resp.AddMessageWithType(constants.CodeOrdersListed, constants.GetDescription(constants.CodeOrdersListed), constants.MessageTypeSuccess)
	return resp, nil
}

var _ mediator.RequestHandler[queries.ListMyOrdersQuery, responses.OrderListResponse] = (*ListMyOrdersHandler)(nil)
