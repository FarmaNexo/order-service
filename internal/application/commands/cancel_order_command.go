package commands

type CancelOrderCommand struct {
	UserID  string `json:"user_id"`
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

func (c CancelOrderCommand) GetName() string { return "CancelOrderCommand" }
