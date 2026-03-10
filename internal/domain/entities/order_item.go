package entities

import "encoding/json"

type OrderItem struct {
	ID              string          `gorm:"column:id;type:uuid;primaryKey"`
	OrderID         string          `gorm:"column:order_id;type:uuid;not null"`
	ProductID       string          `gorm:"column:product_id;type:uuid;not null"`
	ProductSnapshot json.RawMessage `gorm:"column:product_snapshot;type:jsonb;not null"`
	Quantity        int             `gorm:"column:quantity;not null"`
	UnitPrice       float64         `gorm:"column:unit_price;type:decimal(10,2);not null"`
	Subtotal        float64         `gorm:"column:subtotal;type:decimal(10,2);not null"`
}

func (OrderItem) TableName() string {
	return `"orders".order_items`
}

// ProductSnapshotData estructura del snapshot del producto
type ProductSnapshotData struct {
	ProductID        string  `json:"product_id"`
	ProductName      string  `json:"product_name"`
	Description      string  `json:"description,omitempty"`
	ActiveIngredient string  `json:"active_ingredient,omitempty"`
	Presentation     string  `json:"presentation,omitempty"`
	ImageURL         string  `json:"image_url,omitempty"`
	PharmacyID       string  `json:"pharmacy_id"`
	PharmacyName     string  `json:"pharmacy_name"`
	UnitPrice        float64 `json:"unit_price"`
}
