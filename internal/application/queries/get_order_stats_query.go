package queries

type GetOrderStatsQuery struct {
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
}

func (q GetOrderStatsQuery) GetName() string { return "GetOrderStatsQuery" }
