package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/domain/entities"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/pkg/mediator"
)

type UpdateOrderStatusValidator struct{}

func NewUpdateOrderStatusValidator() *UpdateOrderStatusValidator {
	return &UpdateOrderStatusValidator{}
}

func (v *UpdateOrderStatusValidator) Validate(ctx context.Context, cmd commands.UpdateOrderStatusCommand) error {
	var errors []string

	if strings.TrimSpace(cmd.OrderID) == "" {
		errors = append(errors, "El ID de la orden es requerido")
	}

	validStatuses := map[string]bool{
		string(entities.OrderStatusConfirmed):  true,
		string(entities.OrderStatusPreparing):  true,
		string(entities.OrderStatusReady):      true,
		string(entities.OrderStatusInDelivery): true,
		string(entities.OrderStatusCompleted):  true,
	}
	if !validStatuses[cmd.Status] {
		errors = append(errors, "Estado inválido. Use: confirmed, preparing, ready, in_delivery o completed")
	}

	if len(errors) > 0 {
		return fmt.Errorf("errores de validación: %s", strings.Join(errors, "; "))
	}
	return nil
}

var _ mediator.Validator[commands.UpdateOrderStatusCommand, responses.OrderDetailResponse] = (*UpdateOrderStatusValidator)(nil)
