package requests

type AddCartItemRequest struct {
	ProductID  string `json:"product_id" example:"uuid"`
	PharmacyID string `json:"pharmacy_id" example:"uuid"`
	Quantity   int    `json:"quantity" example:"2"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" example:"3"`
}

type CheckoutRequest struct {
	DeliveryMethod    string                `json:"delivery_method" example:"delivery"`
	DeliveryAddressID string                `json:"delivery_address_id,omitempty" example:"uuid"`
	PaymentMethod     string                `json:"payment_method" example:"card"`
	PaymentDetails    PaymentDetailsRequest `json:"payment_details,omitempty"`
	Notes             string                `json:"notes,omitempty" example:"Entregar en recepción"`
}

type PaymentDetailsRequest struct {
	CardToken string `json:"card_token,omitempty" example:"tok_xxx"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason" example:"Cambié de opinión"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" example:"preparing"`
	Notes  string `json:"notes,omitempty" example:"Pedido listo para recoger"`
}
