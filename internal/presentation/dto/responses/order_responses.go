package responses

import "time"

// Cart responses
type CartResponse struct {
	UserID            string                  `json:"user_id"`
	Items             []CartItemResponse      `json:"items"`
	GroupedByPharmacy []PharmacyGroupResponse `json:"grouped_by_pharmacy"`
	TotalItems        int                     `json:"total_items"`
	TotalAmount       float64                 `json:"total_amount"`
	UpdatedAt         time.Time               `json:"updated_at"`
}

type CartItemResponse struct {
	ID             string  `json:"id"`
	ProductID      string  `json:"product_id"`
	ProductName    string  `json:"product_name"`
	PharmacyID     string  `json:"pharmacy_id"`
	PharmacyName   string  `json:"pharmacy_name"`
	Quantity       int     `json:"quantity"`
	UnitPrice      float64 `json:"unit_price"`
	Subtotal       float64 `json:"subtotal"`
	StockAvailable int     `json:"stock_available"`
}

type PharmacyGroupResponse struct {
	PharmacyID   string             `json:"pharmacy_id"`
	PharmacyName string             `json:"pharmacy_name"`
	Items        []CartItemResponse `json:"items"`
	Subtotal     float64            `json:"subtotal"`
	ItemsCount   int                `json:"items_count"`
}

// Checkout responses
type CheckoutResponse struct {
	Orders          []CheckoutOrderResponse `json:"orders"`
	TotalAmount     float64                 `json:"total_amount"`
	PaymentRequired bool                    `json:"payment_required"`
}

type CheckoutOrderResponse struct {
	OrderID      string  `json:"order_id"`
	OrderNumber  string  `json:"order_number"`
	PharmacyID   string  `json:"pharmacy_id"`
	PharmacyName string  `json:"pharmacy_name"`
	Subtotal     float64 `json:"subtotal"`
	DeliveryFee  float64 `json:"delivery_fee"`
	Total        float64 `json:"total"`
	Status       string  `json:"status"`
}

// Order responses
type OrderDetailResponse struct {
	OrderID         string                   `json:"order_id"`
	OrderNumber     string                   `json:"order_number"`
	UserID          string                   `json:"user_id"`
	PharmacyID      string                   `json:"pharmacy_id"`
	PharmacyName    string                   `json:"pharmacy_name"`
	Items           []OrderItemResponse      `json:"items"`
	Subtotal        float64                  `json:"subtotal"`
	DeliveryFee     float64                  `json:"delivery_fee"`
	Total           float64                  `json:"total"`
	DeliveryMethod  string                   `json:"delivery_method"`
	DeliveryAddress *DeliveryAddressResponse `json:"delivery_address,omitempty"`
	PaymentMethod   string                   `json:"payment_method"`
	PaymentStatus   string                   `json:"payment_status"`
	Status          string                   `json:"status"`
	StatusHistory   []StatusHistoryResponse  `json:"status_history"`
	Notes           string                   `json:"notes,omitempty"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
}

type OrderItemResponse struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
}

type DeliveryAddressResponse struct {
	Street    string `json:"street"`
	City      string `json:"city"`
	State     string `json:"state"`
	Reference string `json:"reference,omitempty"`
}

type StatusHistoryResponse struct {
	Status    string    `json:"status"`
	Notes     string    `json:"notes,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Order list response
type OrderListResponse struct {
	Orders     []OrderSummaryResponse `json:"orders"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}

type OrderSummaryResponse struct {
	OrderID      string    `json:"order_id"`
	OrderNumber  string    `json:"order_number"`
	PharmacyID   string    `json:"pharmacy_id"`
	PharmacyName string    `json:"pharmacy_name"`
	Total        float64   `json:"total"`
	Status       string    `json:"status"`
	ItemsCount   int       `json:"items_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// Empty response for clear cart
type EmptyResponse struct{}

// Stats response
type OrderStatsResponse struct {
	TotalOrders   int64                 `json:"total_orders"`
	TotalRevenue  float64               `json:"total_revenue"`
	ByStatus      map[string]int64      `json:"by_status"`
	TopPharmacies []TopPharmacyResponse `json:"top_pharmacies"`
	AvgOrderValue float64               `json:"avg_order_value"`
}

type TopPharmacyResponse struct {
	PharmacyID   string  `json:"pharmacy_id"`
	PharmacyName string  `json:"pharmacy_name"`
	OrdersCount  int64   `json:"orders_count"`
	Revenue      float64 `json:"revenue"`
}
