package events

import "time"

const (
	EventOrderCreated       = "ORDER_CREATED"
	EventOrderConfirmed     = "ORDER_CONFIRMED"
	EventOrderStatusUpdated = "ORDER_STATUS_UPDATED"
	EventOrderCompleted     = "ORDER_COMPLETED"
	EventOrderCancelled     = "ORDER_CANCELLED"
	EventPaymentCompleted   = "PAYMENT_COMPLETED"
	EventPaymentFailed      = "PAYMENT_FAILED"
)

type OrderEvent struct {
	EventType   string            `json:"event_type"`
	OrderID     string            `json:"order_id,omitempty"`
	OrderNumber string            `json:"order_number,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	PharmacyID  string            `json:"pharmacy_id,omitempty"`
	Status      string            `json:"status,omitempty"`
	Total       float64           `json:"total,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	Metadata    map[string]string `json:"metadata"`
}

func NewOrderEvent(eventType string) OrderEvent {
	return OrderEvent{
		EventType: eventType,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"source":  "order-service",
			"version": "1.0",
		},
	}
}

func (e OrderEvent) WithOrder(orderID, orderNumber string) OrderEvent {
	e.OrderID = orderID
	e.OrderNumber = orderNumber
	return e
}

func (e OrderEvent) WithUser(userID string) OrderEvent {
	e.UserID = userID
	return e
}

func (e OrderEvent) WithPharmacy(pharmacyID string) OrderEvent {
	e.PharmacyID = pharmacyID
	return e
}

func (e OrderEvent) WithStatus(status string) OrderEvent {
	e.Status = status
	return e
}

func (e OrderEvent) WithTotal(total float64) OrderEvent {
	e.Total = total
	return e
}

// InventoryUpdatedEvent evento consumido de Pharmacy Service
type InventoryUpdatedEvent struct {
	EventType  string                       `json:"event_type"`
	ProductID  string                       `json:"product_id"`
	PharmacyID string                       `json:"pharmacy_id"`
	Stock      int                          `json:"stock"`
	Price      float64                      `json:"price"`
	Timestamp  time.Time                    `json:"timestamp"`
	Metadata   InventoryUpdatedEventMetadata `json:"metadata"`
}

type InventoryUpdatedEventMetadata struct {
	PharmacyName string `json:"pharmacy_name"`
	ProductName  string `json:"product_name"`
}

// ProductUpdatedEvent evento consumido de Catalog Service
type ProductUpdatedEvent struct {
	EventType string    `json:"event_type"`
	ProductID string    `json:"product_id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}
