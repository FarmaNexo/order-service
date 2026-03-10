package handlers

import (
	"encoding/json"

	"github.com/farmanexo/order-service/internal/domain/entities"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
)

func buildOrderDetail(order *entities.Order) responses.OrderDetailResponse {
	items := make([]responses.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		var snapshot entities.ProductSnapshotData
		json.Unmarshal(item.ProductSnapshot, &snapshot)

		items = append(items, responses.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: snapshot.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    item.Subtotal,
		})
	}

	statusHistory := make([]responses.StatusHistoryResponse, 0, len(order.StatusHistory))
	for _, sh := range order.StatusHistory {
		notes := ""
		if sh.Notes != nil {
			notes = *sh.Notes
		}
		statusHistory = append(statusHistory, responses.StatusHistoryResponse{
			Status:    sh.Status,
			Notes:     notes,
			Timestamp: sh.CreatedAt,
		})
	}

	var deliveryAddr *responses.DeliveryAddressResponse
	if order.DeliveryAddressSnapshot != nil {
		var addr responses.DeliveryAddressResponse
		if json.Unmarshal(order.DeliveryAddressSnapshot, &addr) == nil {
			deliveryAddr = &addr
		}
	}

	notes := ""
	if order.Notes != nil {
		notes = *order.Notes
	}

	return responses.OrderDetailResponse{
		OrderID:         order.ID,
		OrderNumber:     order.OrderNumber,
		UserID:          order.UserID,
		PharmacyID:      order.PharmacyID,
		Items:           items,
		Subtotal:        order.Subtotal,
		DeliveryFee:     order.DeliveryFee,
		Total:           order.Total,
		DeliveryMethod:  order.DeliveryMethod,
		DeliveryAddress: deliveryAddr,
		PaymentMethod:   order.PaymentMethod,
		PaymentStatus:   order.PaymentStatus,
		Status:          order.Status,
		StatusHistory:   statusHistory,
		Notes:           notes,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}
