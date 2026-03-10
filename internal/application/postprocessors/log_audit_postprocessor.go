package postprocessors

import (
	"context"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/application/queries"
	"github.com/farmanexo/order-service/pkg/mediator"
	"go.uber.org/zap"
)

type LogAuditPostProcessor struct {
	logger *zap.Logger
}

func NewLogAuditPostProcessor(logger *zap.Logger) *LogAuditPostProcessor {
	return &LogAuditPostProcessor{logger: logger}
}

func (p *LogAuditPostProcessor) Process(ctx context.Context, request interface{}, response interface{}) error {
	userID := p.getUserIDFromContext(ctx)
	correlationID := mediator.GetCorrelationID(ctx)
	isSuccess := p.checkSuccess(response)

	switch request.(type) {
	case commands.AddCartItemCommand, *commands.AddCartItemCommand:
		p.logAudit("CART_ITEM_ADDED", userID, correlationID, isSuccess)
	case commands.UpdateCartItemCommand, *commands.UpdateCartItemCommand:
		p.logAudit("CART_ITEM_UPDATED", userID, correlationID, isSuccess)
	case commands.RemoveCartItemCommand, *commands.RemoveCartItemCommand:
		p.logAudit("CART_ITEM_REMOVED", userID, correlationID, isSuccess)
	case commands.ClearCartCommand, *commands.ClearCartCommand:
		p.logAudit("CART_CLEARED", userID, correlationID, isSuccess)
	case commands.CheckoutCommand, *commands.CheckoutCommand:
		p.logAudit("CHECKOUT_COMPLETED", userID, correlationID, isSuccess)
	case commands.CancelOrderCommand, *commands.CancelOrderCommand:
		p.logAudit("ORDER_CANCELLED", userID, correlationID, isSuccess)
	case commands.UpdateOrderStatusCommand, *commands.UpdateOrderStatusCommand:
		p.logAudit("ORDER_STATUS_UPDATED", userID, correlationID, isSuccess)
	case queries.GetCartQuery, *queries.GetCartQuery:
		p.logAudit("CART_VIEWED", userID, correlationID, isSuccess)
	case queries.ListMyOrdersQuery, *queries.ListMyOrdersQuery:
		p.logAudit("ORDERS_LISTED", userID, correlationID, isSuccess)
	case queries.GetOrderDetailQuery, *queries.GetOrderDetailQuery:
		p.logAudit("ORDER_DETAIL_VIEWED", userID, correlationID, isSuccess)
	}

	return nil
}

func (p *LogAuditPostProcessor) getUserIDFromContext(ctx context.Context) string {
	userID, ok := mediator.GetUserID(ctx)
	if !ok {
		return "anonymous"
	}
	return userID
}

func (p *LogAuditPostProcessor) checkSuccess(response interface{}) bool {
	if resp, ok := response.(interface{ IsValid() bool }); ok {
		return resp.IsValid()
	}
	return true
}

func (p *LogAuditPostProcessor) logAudit(action, userID, correlationID string, success bool) {
	p.logger.Info("Audit",
		zap.String("action", action),
		zap.String("user_id", userID),
		zap.String("correlation_id", correlationID),
		zap.Bool("success", success),
	)
}
