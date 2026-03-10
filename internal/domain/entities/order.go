package entities

import (
	"encoding/json"
	"time"
)

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusConfirmed      OrderStatus = "confirmed"
	OrderStatusPreparing      OrderStatus = "preparing"
	OrderStatusReady          OrderStatus = "ready"
	OrderStatusInDelivery     OrderStatus = "in_delivery"
	OrderStatusCompleted      OrderStatus = "completed"
	OrderStatusCancelled      OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type DeliveryMethod string

const (
	DeliveryMethodDelivery DeliveryMethod = "delivery"
	DeliveryMethodPickup   DeliveryMethod = "pickup"
)

type PaymentMethod string

const (
	PaymentMethodCard PaymentMethod = "card"
	PaymentMethodCash PaymentMethod = "cash"
	PaymentMethodYape PaymentMethod = "yape"
	PaymentMethodPlin PaymentMethod = "plin"
)

type Order struct {
	ID                      string          `gorm:"column:id;type:uuid;primaryKey"`
	OrderNumber             string          `gorm:"column:order_number;type:varchar(50);uniqueIndex;not null"`
	UserID                  string          `gorm:"column:user_id;type:uuid;not null"`
	PharmacyID              string          `gorm:"column:pharmacy_id;type:uuid;not null"`
	Subtotal                float64         `gorm:"column:subtotal;type:decimal(10,2);not null"`
	DeliveryFee             float64         `gorm:"column:delivery_fee;type:decimal(10,2);not null;default:0"`
	Total                   float64         `gorm:"column:total;type:decimal(10,2);not null"`
	DeliveryMethod          string          `gorm:"column:delivery_method;type:varchar(20);not null"`
	DeliveryAddressID       *string         `gorm:"column:delivery_address_id;type:uuid"`
	DeliveryAddressSnapshot json.RawMessage `gorm:"column:delivery_address_snapshot;type:jsonb"`
	PaymentMethod           string          `gorm:"column:payment_method;type:varchar(20);not null"`
	PaymentStatus           string          `gorm:"column:payment_status;type:varchar(20);not null;default:pending"`
	PaymentTransactionID    *string         `gorm:"column:payment_transaction_id;type:varchar(100)"`
	Status                  string          `gorm:"column:status;type:varchar(20);not null;default:pending_payment"`
	Notes                   *string         `gorm:"column:notes;type:text"`
	CancellationReason      *string         `gorm:"column:cancellation_reason;type:text"`
	CreatedAt               time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt               time.Time       `gorm:"column:updated_at;autoUpdateTime"`
	CompletedAt             *time.Time      `gorm:"column:completed_at"`
	CancelledAt             *time.Time      `gorm:"column:cancelled_at"`

	// Relaciones
	Items         []OrderItem          `gorm:"foreignKey:OrderID"`
	StatusHistory []OrderStatusHistory `gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string {
	return `"orders".orders`
}

// ValidStatusTransitions define las transiciones válidas de estado
var ValidStatusTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPendingPayment: {OrderStatusConfirmed, OrderStatusCancelled},
	OrderStatusConfirmed:      {OrderStatusPreparing, OrderStatusCancelled},
	OrderStatusPreparing:      {OrderStatusReady},
	OrderStatusReady:          {OrderStatusInDelivery, OrderStatusCompleted},
	OrderStatusInDelivery:     {OrderStatusCompleted},
}

// CanTransitionTo verifica si una transición de estado es válida
func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	currentStatus := OrderStatus(o.Status)
	validTransitions, exists := ValidStatusTransitions[currentStatus]
	if !exists {
		return false
	}
	for _, valid := range validTransitions {
		if valid == newStatus {
			return true
		}
	}
	return false
}

// IsCancellable verifica si la orden se puede cancelar
func (o *Order) IsCancellable() bool {
	return o.Status == string(OrderStatusPendingPayment) || o.Status == string(OrderStatusConfirmed)
}
