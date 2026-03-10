package services

import (
	"context"

	"github.com/farmanexo/order-service/internal/domain/events"
)

type EventPublisher interface {
	Publish(ctx context.Context, event events.OrderEvent) error
}
