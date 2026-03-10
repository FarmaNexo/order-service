package commands

type RemoveCartItemCommand struct {
	UserID string `json:"user_id"`
	ItemID string `json:"item_id"`
}

func (c RemoveCartItemCommand) GetName() string { return "RemoveCartItemCommand" }
