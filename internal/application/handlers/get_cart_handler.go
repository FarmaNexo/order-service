package handlers

import (
	"context"
	"time"

	"github.com/farmanexo/order-service/internal/application/queries"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/internal/shared/constants"
	"github.com/farmanexo/order-service/pkg/mediator"
	"go.uber.org/zap"
)

type GetCartHandler struct {
	cartRepo repositories.CartRepository
	logger   *zap.Logger
}

func NewGetCartHandler(cartRepo repositories.CartRepository, logger *zap.Logger) *GetCartHandler {
	return &GetCartHandler{cartRepo: cartRepo, logger: logger}
}

func (h *GetCartHandler) Handle(ctx context.Context, query queries.GetCartQuery) (*common.ApiResponse[responses.CartResponse], error) {
	items, err := h.cartRepo.FindByUserID(ctx, query.UserID)
	if err != nil {
		h.logger.Error("Error obteniendo carrito", zap.Error(err))
		return common.InternalServerErrorResponse[responses.CartResponse]("Error obteniendo carrito"), nil
	}

	// Build cart items
	cartItems := make([]responses.CartItemResponse, 0, len(items))
	for _, item := range items {
		cartItems = append(cartItems, responses.CartItemResponse{
			ID:           item.ID,
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			PharmacyID:   item.PharmacyID,
			PharmacyName: item.PharmacyName,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			Subtotal:     item.Subtotal(),
		})
	}

	// Group by pharmacy
	pharmacyMap := make(map[string]*responses.PharmacyGroupResponse)
	for _, item := range cartItems {
		group, exists := pharmacyMap[item.PharmacyID]
		if !exists {
			group = &responses.PharmacyGroupResponse{
				PharmacyID:   item.PharmacyID,
				PharmacyName: item.PharmacyName,
				Items:        make([]responses.CartItemResponse, 0),
			}
			pharmacyMap[item.PharmacyID] = group
		}
		group.Items = append(group.Items, item)
		group.Subtotal += item.Subtotal
		group.ItemsCount++
	}

	groups := make([]responses.PharmacyGroupResponse, 0, len(pharmacyMap))
	for _, group := range pharmacyMap {
		groups = append(groups, *group)
	}

	var totalAmount float64
	for _, item := range cartItems {
		totalAmount += item.Subtotal
	}

	cartResponse := responses.CartResponse{
		UserID:            query.UserID,
		Items:             cartItems,
		GroupedByPharmacy: groups,
		TotalItems:        len(cartItems),
		TotalAmount:       totalAmount,
		UpdatedAt:         time.Now(),
	}

	resp := common.OkResponse(cartResponse)
	resp.AddMessageWithType(constants.CodeCartRetrieved, constants.GetDescription(constants.CodeCartRetrieved), constants.MessageTypeSuccess)
	return resp, nil
}

var _ mediator.RequestHandler[queries.GetCartQuery, responses.CartResponse] = (*GetCartHandler)(nil)
