package queries

type ListPharmacyOrdersQuery struct {
	UserID     string `json:"user_id"`
	PharmacyID string `json:"pharmacy_id"`
	Status     string `json:"status"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

func (q ListPharmacyOrdersQuery) GetName() string { return "ListPharmacyOrdersQuery" }
