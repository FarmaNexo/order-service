package commands

type UpdateOrderStatusCommand struct {
	UserID     string `json:"user_id"`
	OrderID    string `json:"order_id"`
	Status     string `json:"status"`
	Notes      string `json:"notes"`
	PharmacyID string `json:"pharmacy_id"` // resolved from user's pharmacy
}

func (c UpdateOrderStatusCommand) GetName() string { return "UpdateOrderStatusCommand" }
