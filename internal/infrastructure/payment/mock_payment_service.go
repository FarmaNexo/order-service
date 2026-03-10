package payment

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/farmanexo/order-service/internal/domain/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MockPaymentService struct {
	successRate float64
	logger      *zap.Logger
}

func NewMockPaymentService(successRate float64, logger *zap.Logger) *MockPaymentService {
	return &MockPaymentService{successRate: successRate, logger: logger}
}

func (s *MockPaymentService) ProcessPayment(ctx context.Context, request services.PaymentRequest) (*services.PaymentResult, error) {
	s.logger.Info("Procesando pago (mock)",
		zap.Float64("amount", request.Amount),
		zap.String("method", request.PaymentMethod),
		zap.String("user_id", request.UserID),
	)

	// Simulate processing delay
	time.Sleep(500 * time.Millisecond)

	// Random success based on success rate
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Float64() <= s.successRate {
		return &services.PaymentResult{
			Success:       true,
			TransactionID: fmt.Sprintf("TXN-%s", uuid.New().String()[:8]),
			Message:       "Pago procesado exitosamente",
		}, nil
	}

	return &services.PaymentResult{
		Success:       false,
		TransactionID: "",
		Message:       "Pago rechazado. Intente nuevamente o use otro método de pago",
	}, nil
}

func (s *MockPaymentService) RefundPayment(ctx context.Context, transactionID string, amount float64) (*services.PaymentResult, error) {
	s.logger.Info("Procesando reembolso (mock)",
		zap.String("transaction_id", transactionID),
		zap.Float64("amount", amount),
	)

	time.Sleep(300 * time.Millisecond)

	return &services.PaymentResult{
		Success:       true,
		TransactionID: fmt.Sprintf("RFN-%s", uuid.New().String()[:8]),
		Message:       "Reembolso procesado exitosamente",
	}, nil
}

var _ services.PaymentService = (*MockPaymentService)(nil)
