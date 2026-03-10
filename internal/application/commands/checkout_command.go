package commands

type CheckoutCommand struct {
	UserID            string         `json:"user_id"`
	DeliveryMethod    string         `json:"delivery_method"`
	DeliveryAddressID string         `json:"delivery_address_id"`
	PaymentMethod     string         `json:"payment_method"`
	PaymentDetails    PaymentDetails `json:"payment_details"`
	Notes             string         `json:"notes"`
	AccessToken       string         `json:"-"`
}

type PaymentDetails struct {
	CardToken string `json:"card_token,omitempty"`
}

func (c CheckoutCommand) GetName() string { return "CheckoutCommand" }
