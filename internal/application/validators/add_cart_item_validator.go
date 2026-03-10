package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/pkg/mediator"
)

type AddCartItemValidator struct{}

func NewAddCartItemValidator() *AddCartItemValidator {
	return &AddCartItemValidator{}
}

func (v *AddCartItemValidator) Validate(ctx context.Context, cmd commands.AddCartItemCommand) error {
	var errors []string

	if strings.TrimSpace(cmd.ProductID) == "" {
		errors = append(errors, "El ID del producto es requerido")
	}
	if strings.TrimSpace(cmd.PharmacyID) == "" {
		errors = append(errors, "El ID de la farmacia es requerido")
	}
	if cmd.Quantity <= 0 {
		errors = append(errors, "La cantidad debe ser mayor a 0")
	}
	if cmd.Quantity > 100 {
		errors = append(errors, "La cantidad máxima por producto es 100")
	}

	if len(errors) > 0 {
		return fmt.Errorf("errores de validación: %s", strings.Join(errors, "; "))
	}
	return nil
}

var _ mediator.Validator[commands.AddCartItemCommand, responses.CartResponse] = (*AddCartItemValidator)(nil)
