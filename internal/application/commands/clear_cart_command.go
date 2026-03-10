package commands

type ClearCartCommand struct {
	UserID string `json:"user_id"`
}

func (c ClearCartCommand) GetName() string { return "ClearCartCommand" }
