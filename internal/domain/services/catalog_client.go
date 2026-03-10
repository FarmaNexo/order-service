package services

import "context"

type ProductInfo struct {
	ProductID        string  `json:"product_id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	ActiveIngredient string  `json:"active_ingredient"`
	Presentation     string  `json:"presentation"`
	ImageURL         string  `json:"image_url"`
	Price            float64 `json:"price"`
}

type CatalogClient interface {
	GetProduct(ctx context.Context, productID string) (*ProductInfo, error)
}
