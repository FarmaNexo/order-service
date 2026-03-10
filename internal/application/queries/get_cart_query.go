package queries

type GetCartQuery struct {
	UserID string `json:"user_id"`
}

func (q GetCartQuery) GetName() string { return "GetCartQuery" }
