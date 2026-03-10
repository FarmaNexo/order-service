package commands

type AddCartItemCommand struct {
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	PharmacyID string `json:"pharmacy_id"`
	Quantity   int    `json:"quantity"`
}

func (c AddCartItemCommand) GetName() string { return "AddCartItemCommand" }
