package queries

type ListMyOrdersQuery struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

func (q ListMyOrdersQuery) GetName() string { return "ListMyOrdersQuery" }
