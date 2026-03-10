package preprocessors

import (
	"context"
	"strings"

	"github.com/farmanexo/order-service/internal/application/commands"
	"go.uber.org/zap"
)

type SanitizeInputPreProcessor struct {
	logger *zap.Logger
}

func NewSanitizeInputPreProcessor(logger *zap.Logger) *SanitizeInputPreProcessor {
	return &SanitizeInputPreProcessor{logger: logger}
}

func (p *SanitizeInputPreProcessor) Process(ctx context.Context, request interface{}) error {
	switch cmd := request.(type) {
	case *commands.AddCartItemCommand:
		cmd.ProductID = strings.TrimSpace(cmd.ProductID)
		cmd.PharmacyID = strings.TrimSpace(cmd.PharmacyID)
	case *commands.CheckoutCommand:
		cmd.DeliveryMethod = strings.TrimSpace(strings.ToLower(cmd.DeliveryMethod))
		cmd.PaymentMethod = strings.TrimSpace(strings.ToLower(cmd.PaymentMethod))
		cmd.DeliveryAddressID = strings.TrimSpace(cmd.DeliveryAddressID)
		cmd.Notes = strings.TrimSpace(cmd.Notes)
	case *commands.CancelOrderCommand:
		cmd.OrderID = strings.TrimSpace(cmd.OrderID)
		cmd.Reason = strings.TrimSpace(cmd.Reason)
	case *commands.UpdateOrderStatusCommand:
		cmd.OrderID = strings.TrimSpace(cmd.OrderID)
		cmd.Status = strings.TrimSpace(strings.ToLower(cmd.Status))
		cmd.Notes = strings.TrimSpace(cmd.Notes)
	}
	return nil
}
