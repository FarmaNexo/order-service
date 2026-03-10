package queries

type ListAllOrdersQuery struct {
	UserID     string `json:"user_id"`
	PharmacyID string `json:"pharmacy_id"`
	Status     string `json:"status"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

func (q ListAllOrdersQuery) GetName() string { return "ListAllOrdersQuery" }
