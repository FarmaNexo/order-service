package commands

type UpdateCartItemCommand struct {
	UserID   string `json:"user_id"`
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

func (c UpdateCartItemCommand) GetName() string { return "UpdateCartItemCommand" }
