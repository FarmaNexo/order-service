package services

import "context"

type PaymentRequest struct {
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	CardToken     string  `json:"card_token,omitempty"`
	Description   string  `json:"description"`
	UserID        string  `json:"user_id"`
}

type PaymentResult struct {
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id"`
	Message       string `json:"message"`
}

type PaymentService interface {
	ProcessPayment(ctx context.Context, request PaymentRequest) (*PaymentResult, error)
	RefundPayment(ctx context.Context, transactionID string, amount float64) (*PaymentResult, error)
}
