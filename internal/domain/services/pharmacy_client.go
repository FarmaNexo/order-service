package services

import "context"

type PharmacyInventoryItem struct {
	PharmacyID   string  `json:"pharmacy_id"`
	PharmacyName string  `json:"pharmacy_name"`
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Stock        int     `json:"stock"`
	Price        float64 `json:"price"`
	IsAvailable  bool    `json:"is_available"`
}

type PharmacyInfo struct {
	PharmacyID string `json:"pharmacy_id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
}

type PharmacyClient interface {
	GetInventoryItem(ctx context.Context, pharmacyID, productID string) (*PharmacyInventoryItem, error)
	GetPharmacyInfo(ctx context.Context, pharmacyID string) (*PharmacyInfo, error)
}
