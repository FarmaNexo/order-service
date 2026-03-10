package queries

type GetOrderDetailQuery struct {
	UserID  string `json:"user_id"`
	OrderID string `json:"order_id"`
}

func (q GetOrderDetailQuery) GetName() string { return "GetOrderDetailQuery" }
