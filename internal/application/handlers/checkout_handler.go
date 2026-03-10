package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/domain/entities"
	"github.com/farmanexo/order-service/internal/domain/events"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/farmanexo/order-service/internal/domain/services"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/internal/shared/constants"
	"github.com/farmanexo/order-service/pkg/config"
	"github.com/farmanexo/order-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CheckoutHandler struct {
	cartRepo       repositories.CartRepository
	orderRepo      repositories.OrderRepository
	orderItemRepo  repositories.OrderItemRepository
	statusHistRepo repositories.OrderStatusHistoryRepository
	pharmacyClient services.PharmacyClient
	userClient     services.UserClient
	paymentService services.PaymentService
	eventPublisher services.EventPublisher
	deliveryCfg    config.DeliveryConfig
	logger         *zap.Logger
}

func NewCheckoutHandler(
	cartRepo repositories.CartRepository,
	orderRepo repositories.OrderRepository,
	orderItemRepo repositories.OrderItemRepository,
	statusHistRepo repositories.OrderStatusHistoryRepository,
	pharmacyClient services.PharmacyClient,
	userClient services.UserClient,
	paymentService services.PaymentService,
	eventPublisher services.EventPublisher,
	deliveryCfg config.DeliveryConfig,
	logger *zap.Logger,
) *CheckoutHandler {
	return &CheckoutHandler{
		cartRepo: cartRepo, orderRepo: orderRepo, orderItemRepo: orderItemRepo,
		statusHistRepo: statusHistRepo, pharmacyClient: pharmacyClient,
		userClient: userClient, paymentService: paymentService,
		eventPublisher: eventPublisher, deliveryCfg: deliveryCfg, logger: logger,
	}
}

func (h *CheckoutHandler) Handle(ctx context.Context, cmd commands.CheckoutCommand) (*common.ApiResponse[responses.CheckoutResponse], error) {
	// 1. Get cart items
	cartItems, err := h.cartRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		h.logger.Error("Error obteniendo carrito", zap.Error(err))
		return common.InternalServerErrorResponse[responses.CheckoutResponse]("Error obteniendo carrito"), nil
	}

	if len(cartItems) == 0 {
		return common.BadRequestResponse[responses.CheckoutResponse](constants.CodeCartEmpty, "El carrito está vacío"), nil
	}

	// 2. Validate delivery address if delivery method
	var addressSnapshot json.RawMessage
	if cmd.DeliveryMethod == "delivery" {
		if cmd.DeliveryAddressID == "" {
			return common.BadRequestResponse[responses.CheckoutResponse](constants.CodeRequiredField, "La dirección de entrega es requerida para delivery"), nil
		}
		address, err := h.userClient.GetAddress(ctx, cmd.UserID, cmd.DeliveryAddressID, cmd.AccessToken)
		if err != nil || address == nil {
			return common.BadRequestResponse[responses.CheckoutResponse](constants.CodeAddressNotFound, "Dirección de entrega no encontrada"), nil
		}
		addrJSON, _ := json.Marshal(address)
		addressSnapshot = addrJSON
	}

	// 3. Group cart items by pharmacy
	pharmacyGroups := make(map[string][]entities.CartItem)
	for _, item := range cartItems {
		pharmacyGroups[item.PharmacyID] = append(pharmacyGroups[item.PharmacyID], item)
	}

	// 4. Create orders
	var createdOrders []responses.CheckoutOrderResponse
	var totalAmount float64
	var orderEntities []*entities.Order

	for pharmacyID, items := range pharmacyGroups {
		// Calculate subtotal
		var subtotal float64
		for _, item := range items {
			subtotal += item.Subtotal()
		}

		// Calculate delivery fee
		deliveryFee := h.calculateDeliveryFee(subtotal, cmd.DeliveryMethod)

		// Generate order number
		seqNum, err := h.orderRepo.GetNextOrderNumber(ctx)
		if err != nil {
			h.logger.Error("Error generando número de orden", zap.Error(err))
			return common.InternalServerErrorResponse[responses.CheckoutResponse]("Error generando número de orden"), nil
		}
		orderNumber := fmt.Sprintf("ORD-%d-%06d", time.Now().Year(), seqNum)

		orderTotal := subtotal + deliveryFee
		notes := cmd.Notes
		var notesPtr *string
		if notes != "" {
			notesPtr = &notes
		}

		var addrIDPtr *string
		if cmd.DeliveryAddressID != "" {
			addrIDPtr = &cmd.DeliveryAddressID
		}

		order := &entities.Order{
			ID:                      uuid.New().String(),
			OrderNumber:             orderNumber,
			UserID:                  cmd.UserID,
			PharmacyID:              pharmacyID,
			Subtotal:                subtotal,
			DeliveryFee:             deliveryFee,
			Total:                   orderTotal,
			DeliveryMethod:          cmd.DeliveryMethod,
			DeliveryAddressID:       addrIDPtr,
			DeliveryAddressSnapshot: addressSnapshot,
			PaymentMethod:           cmd.PaymentMethod,
			PaymentStatus:           string(entities.PaymentStatusPending),
			Status:                  string(entities.OrderStatusPendingPayment),
			Notes:                   notesPtr,
		}

		if err := h.orderRepo.Create(ctx, order); err != nil {
			h.logger.Error("Error creando orden", zap.Error(err))
			return common.InternalServerErrorResponse[responses.CheckoutResponse]("Error creando orden"), nil
		}

		// Create order items
		orderItems := make([]entities.OrderItem, 0, len(items))
		for _, cartItem := range items {
			snapshot := entities.ProductSnapshotData{
				ProductID:    cartItem.ProductID,
				ProductName:  cartItem.ProductName,
				PharmacyID:   cartItem.PharmacyID,
				PharmacyName: cartItem.PharmacyName,
				UnitPrice:    cartItem.UnitPrice,
			}
			snapshotJSON, _ := json.Marshal(snapshot)

			orderItems = append(orderItems, entities.OrderItem{
				ID:              uuid.New().String(),
				OrderID:         order.ID,
				ProductID:       cartItem.ProductID,
				ProductSnapshot: snapshotJSON,
				Quantity:        cartItem.Quantity,
				UnitPrice:       cartItem.UnitPrice,
				Subtotal:        cartItem.Subtotal(),
			})
		}

		if err := h.orderItemRepo.CreateBatch(ctx, orderItems); err != nil {
			h.logger.Error("Error creando items de orden", zap.Error(err))
			return common.InternalServerErrorResponse[responses.CheckoutResponse]("Error creando items de orden"), nil
		}

		// Create initial status history
		h.statusHistRepo.Create(ctx, &entities.OrderStatusHistory{
			ID:      uuid.New().String(),
			OrderID: order.ID,
			Status:  string(entities.OrderStatusPendingPayment),
		})

		orderEntities = append(orderEntities, order)

		var pharmacyName string
		if len(items) > 0 {
			pharmacyName = items[0].PharmacyName
		}

		createdOrders = append(createdOrders, responses.CheckoutOrderResponse{
			OrderID:      order.ID,
			OrderNumber:  orderNumber,
			PharmacyID:   pharmacyID,
			PharmacyName: pharmacyName,
			Subtotal:     subtotal,
			DeliveryFee:  deliveryFee,
			Total:        orderTotal,
			Status:       string(entities.OrderStatusPendingPayment),
		})

		totalAmount += orderTotal
	}

	// 5. Process payment
	paymentResult, err := h.paymentService.ProcessPayment(ctx, services.PaymentRequest{
		Amount:        totalAmount,
		Currency:      "PEN",
		PaymentMethod: cmd.PaymentMethod,
		CardToken:     cmd.PaymentDetails.CardToken,
		Description:   fmt.Sprintf("FarmaNexo - %d orden(es)", len(createdOrders)),
		UserID:        cmd.UserID,
	})

	if err != nil || !paymentResult.Success {
		// Mark orders as payment_failed
		for _, order := range orderEntities {
			order.PaymentStatus = string(entities.PaymentStatusFailed)
			order.Status = string(entities.OrderStatusPendingPayment)
			h.orderRepo.Update(ctx, order)
		}

		msg := "Error procesando el pago"
		if paymentResult != nil {
			msg = paymentResult.Message
		}
		return common.BadRequestResponse[responses.CheckoutResponse](constants.CodePaymentFailed, msg), nil
	}

	// 6. Payment success - update orders
	for i, order := range orderEntities {
		order.PaymentStatus = string(entities.PaymentStatusPaid)
		order.PaymentTransactionID = &paymentResult.TransactionID
		order.Status = string(entities.OrderStatusConfirmed)
		h.orderRepo.Update(ctx, order)

		// Update status history
		h.statusHistRepo.Create(ctx, &entities.OrderStatusHistory{
			ID:      uuid.New().String(),
			OrderID: order.ID,
			Status:  string(entities.OrderStatusConfirmed),
		})

		createdOrders[i].Status = string(entities.OrderStatusConfirmed)

		// Publish event
		go func(o *entities.Order) {
			event := events.NewOrderEvent(events.EventOrderCreated).
				WithOrder(o.ID, o.OrderNumber).
				WithUser(cmd.UserID).
				WithPharmacy(o.PharmacyID).
				WithTotal(o.Total)
			if err := h.eventPublisher.Publish(context.Background(), event); err != nil {
				h.logger.Error("Error publicando evento ORDER_CREATED", zap.Error(err))
			}
		}(order)
	}

	// 7. Clear cart
	h.cartRepo.DeleteByUserID(ctx, cmd.UserID)

	checkoutResp := responses.CheckoutResponse{
		Orders:          createdOrders,
		TotalAmount:     totalAmount,
		PaymentRequired: false,
	}

	resp := common.CreatedResponse(checkoutResp)
	resp.AddMessageWithType(constants.CodeCheckoutCompleted, constants.GetDescription(constants.CodeCheckoutCompleted), constants.MessageTypeSuccess)
	return resp, nil
}

func (h *CheckoutHandler) calculateDeliveryFee(subtotal float64, deliveryMethod string) float64 {
	if deliveryMethod == "pickup" {
		return 0.00
	}
	if subtotal >= h.deliveryCfg.FreeDeliveryThreshold {
		return 0.00
	}
	return h.deliveryCfg.BaseFee
}

var _ mediator.RequestHandler[commands.CheckoutCommand, responses.CheckoutResponse] = (*CheckoutHandler)(nil)
