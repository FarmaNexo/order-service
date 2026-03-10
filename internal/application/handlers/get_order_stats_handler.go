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

type GetOrderStatsHandler struct {
	orderRepo repositories.OrderRepository
	logger    *zap.Logger
}

func NewGetOrderStatsHandler(orderRepo repositories.OrderRepository, logger *zap.Logger) *GetOrderStatsHandler {
	return &GetOrderStatsHandler{orderRepo: orderRepo, logger: logger}
}

func (h *GetOrderStatsHandler) Handle(ctx context.Context, query queries.GetOrderStatsQuery) (*common.ApiResponse[responses.OrderStatsResponse], error) {
	stats, err := h.orderRepo.GetOrderStats(ctx, query.DateFrom, query.DateTo)
	if err != nil {
		h.logger.Error("Error obteniendo estadísticas", zap.Error(err))
		return common.InternalServerErrorResponse[responses.OrderStatsResponse]("Error obteniendo estadísticas"), nil
	}

	topPharmacies := make([]responses.TopPharmacyResponse, 0, len(stats.TopPharmacies))
	for _, p := range stats.TopPharmacies {
		topPharmacies = append(topPharmacies, responses.TopPharmacyResponse{
			PharmacyID:   p.PharmacyID,
			PharmacyName: p.PharmacyName,
			OrdersCount:  p.OrdersCount,
			Revenue:      p.Revenue,
		})
	}

	statsResp := responses.OrderStatsResponse{
		TotalOrders:   stats.TotalOrders,
		TotalRevenue:  stats.TotalRevenue,
		ByStatus:      stats.ByStatus,
		TopPharmacies: topPharmacies,
		AvgOrderValue: stats.AvgOrderValue,
	}

	resp := common.OkResponse(statsResp)
	resp.AddMessageWithType(constants.CodeOrderStatsRetrieved, constants.GetDescription(constants.CodeOrderStatsRetrieved), constants.MessageTypeSuccess)
	return resp, nil
}

var _ mediator.RequestHandler[queries.GetOrderStatsQuery, responses.OrderStatsResponse] = (*GetOrderStatsHandler)(nil)
