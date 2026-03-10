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

type UpdateOrderStatusHandler struct {
	orderRepo      repositories.OrderRepository
	statusHistRepo repositories.OrderStatusHistoryRepository
	eventPublisher services.EventPublisher
	logger         *zap.Logger
}

func NewUpdateOrderStatusHandler(
	orderRepo repositories.OrderRepository,
	statusHistRepo repositories.OrderStatusHistoryRepository,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) *UpdateOrderStatusHandler {
	return &UpdateOrderStatusHandler{
		orderRepo: orderRepo, statusHistRepo: statusHistRepo,
		eventPublisher: eventPublisher, logger: logger,
	}
}

func (h *UpdateOrderStatusHandler) Handle(ctx context.Context, cmd commands.UpdateOrderStatusCommand) (*common.ApiResponse[responses.OrderDetailResponse], error) {
	order, err := h.orderRepo.FindByIDWithItems(ctx, cmd.OrderID)
	if err != nil || order == nil {
		return common.NotFoundResponse[responses.OrderDetailResponse]("Orden no encontrada"), nil
	}

	// Verify pharmacy ownership
	if order.PharmacyID != cmd.PharmacyID {
		return common.ForbiddenResponse[responses.OrderDetailResponse]("No tiene permisos sobre esta orden"), nil
	}

	// Validate status transition
	newStatus := entities.OrderStatus(cmd.Status)
	if !order.CanTransitionTo(newStatus) {
		return common.BadRequestResponse[responses.OrderDetailResponse](constants.CodeInvalidTransition, "Transición de estado no permitida: "+order.Status+" → "+cmd.Status), nil
	}

	// Update order status
	order.Status = cmd.Status
	if cmd.Status == string(entities.OrderStatusCompleted) {
		now := time.Now()
		order.CompletedAt = &now
	}

	if err := h.orderRepo.Update(ctx, order); err != nil {
		h.logger.Error("Error actualizando estado de orden", zap.Error(err))
		return common.InternalServerErrorResponse[responses.OrderDetailResponse]("Error actualizando estado"), nil
	}

	// Status history
	var notesPtr *string
	if cmd.Notes != "" {
		notesPtr = &cmd.Notes
	}
	h.statusHistRepo.Create(ctx, &entities.OrderStatusHistory{
		ID:              uuid.New().String(),
		OrderID:         order.ID,
		Status:          cmd.Status,
		Notes:           notesPtr,
		ChangedByUserID: &cmd.UserID,
	})

	// Publish event
	eventType := events.EventOrderStatusUpdated
	if cmd.Status == string(entities.OrderStatusCompleted) {
		eventType = events.EventOrderCompleted
	}
	go func() {
		event := events.NewOrderEvent(eventType).
			WithOrder(order.ID, order.OrderNumber).
			WithUser(order.UserID).
			WithPharmacy(order.PharmacyID).
			WithStatus(cmd.Status)
		h.eventPublisher.Publish(context.Background(), event)
	}()

	detail := buildOrderDetail(order)
	resp := common.OkResponse(detail)
	resp.AddMessageWithType(constants.CodeOrderStatusUpdated, constants.GetDescription(constants.CodeOrderStatusUpdated), constants.MessageTypeSuccess)
	return resp, nil
}

var _ mediator.RequestHandler[commands.UpdateOrderStatusCommand, responses.OrderDetailResponse] = (*UpdateOrderStatusHandler)(nil)
