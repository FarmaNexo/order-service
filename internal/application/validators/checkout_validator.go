package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/pkg/mediator"
)

type CheckoutValidator struct{}

func NewCheckoutValidator() *CheckoutValidator {
	return &CheckoutValidator{}
}

func (v *CheckoutValidator) Validate(ctx context.Context, cmd commands.CheckoutCommand) error {
	var errors []string

	validDeliveryMethods := map[string]bool{"delivery": true, "pickup": true}
	if !validDeliveryMethods[cmd.DeliveryMethod] {
		errors = append(errors, "Método de entrega inválido. Use: delivery o pickup")
	}

	if cmd.DeliveryMethod == "delivery" && strings.TrimSpace(cmd.DeliveryAddressID) == "" {
		errors = append(errors, "La dirección de entrega es requerida para delivery")
	}

	validPaymentMethods := map[string]bool{"card": true, "cash": true, "yape": true, "plin": true}
	if !validPaymentMethods[cmd.PaymentMethod] {
		errors = append(errors, "Método de pago inválido. Use: card, cash, yape o plin")
	}

	if cmd.PaymentMethod == "card" && strings.TrimSpace(cmd.PaymentDetails.CardToken) == "" {
		errors = append(errors, "El token de tarjeta es requerido para pago con tarjeta")
	}

	if len(errors) > 0 {
		return fmt.Errorf("errores de validación: %s", strings.Join(errors, "; "))
	}
	return nil
}

var _ mediator.Validator[commands.CheckoutCommand, responses.CheckoutResponse] = (*CheckoutValidator)(nil)
