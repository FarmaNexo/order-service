package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/pkg/mediator"
)

type CancelOrderValidator struct{}

func NewCancelOrderValidator() *CancelOrderValidator {
	return &CancelOrderValidator{}
}

func (v *CancelOrderValidator) Validate(ctx context.Context, cmd commands.CancelOrderCommand) error {
	var errors []string

	if strings.TrimSpace(cmd.OrderID) == "" {
		errors = append(errors, "El ID de la orden es requerido")
	}

	if len(errors) > 0 {
		return fmt.Errorf("errores de validación: %s", strings.Join(errors, "; "))
	}
	return nil
}

var _ mediator.Validator[commands.CancelOrderCommand, responses.OrderDetailResponse] = (*CancelOrderValidator)(nil)
