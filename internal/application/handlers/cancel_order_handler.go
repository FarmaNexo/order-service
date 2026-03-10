package handlers

import (
	"context"
	"time"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/domain/entities"
	"github.com/farmanexo/order-service/internal/domain/events"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/farmanexo/order-service/internal/domain/services"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/internal/shared/constants"
	"github.com/farmanexo/order-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CancelOrderHandler struct {
	orderRepo      repositories.OrderRepository
	statusHistRepo repositories.OrderStatusHistoryRepository
	paymentService services.PaymentService
	eventPublisher services.EventPublisher
	logger         *zap.Logger
}

func NewCancelOrderHandler(
	orderRepo repositories.OrderRepository,
	statusHistRepo repositories.OrderStatusHistoryRepository,
	paymentService services.PaymentService,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) *CancelOrderHandler {
	return &CancelOrderHandler{
		orderRepo: orderRepo, statusHistRepo: statusHistRepo,
		paymentService: paymentService, eventPublisher: eventPublisher, logger: logger,
	}
}

func (h *CancelOrderHandler) Handle(ctx context.Context, cmd commands.CancelOrderCommand) (*common.ApiResponse[responses.OrderDetailResponse], error) {
	order, err := h.orderRepo.FindByIDWithItems(ctx, cmd.OrderID)
	if err != nil || order == nil {
		return common.NotFoundResponse[responses.OrderDetailResponse]("Orden no encontrada"), nil
	}

	if order.UserID != cmd.UserID {
		return common.NotFoundResponse[responses.OrderDetailResponse]("Orden no encontrada"), nil
	}

	if !order.IsCancellable() {
		return common.BadRequestResponse[responses.OrderDetailResponse](constants.CodeOrderNotCancellable, "La orden no puede ser cancelada en su estado actual"), nil
	}

	// Cancel order
	now := time.Now()
	order.Status = string(entities.OrderStatusCancelled)
	order.CancellationReason = &cmd.Reason
	order.CancelledAt = &now

	// Refund if paid
	if order.PaymentStatus == string(entities.PaymentStatusPaid) && order.PaymentTransactionID != nil {
		_, refundErr := h.paymentService.RefundPayment(ctx, *order.PaymentTransactionID, order.Total)
		if refundErr != nil {
			h.logger.Error("Error procesando reembolso", zap.Error(refundErr))
		} else {
			order.PaymentStatus = string(entities.PaymentStatusRefunded)
		}
	}

	if err := h.orderRepo.Update(ctx, order); err != nil {
		h.logger.Error("Error cancelando orden", zap.Error(err))
		return common.InternalServerErrorResponse[responses.OrderDetailResponse]("Error cancelando orden"), nil
	}

	// Status history
	h.statusHistRepo.Create(ctx, &entities.OrderStatusHistory{
		ID:              uuid.New().String(),
		OrderID:         order.ID,
		Status:          string(entities.OrderStatusCancelled),
		Notes:           &cmd.Reason,
		ChangedByUserID: &cmd.UserID,
	})

	// Publish event
	go func() {
		event := events.NewOrderEvent(events.EventOrderCancelled).
			WithOrder(order.ID, order.OrderNumber).
			WithUser(cmd.UserID).
			WithPharmacy(order.PharmacyID)
		h.eventPublisher.Publish(context.Background(), event)
	}()

	return h.buildOrderDetailResponse(order), nil
}

func (h *CancelOrderHandler) buildOrderDetailResponse(order *entities.Order) *common.ApiResponse[responses.OrderDetailResponse] {
	detail := buildOrderDetail(order)
	resp := common.OkResponse(detail)
	resp.AddMessageWithType(constants.CodeOrderCancelled, constants.GetDescription(constants.CodeOrderCancelled), constants.MessageTypeSuccess)
	return resp
}

var _ mediator.RequestHandler[commands.CancelOrderCommand, responses.OrderDetailResponse] = (*CancelOrderHandler)(nil)
