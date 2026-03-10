package entities

import "time"

type CartItem struct {
	ID           string    `gorm:"column:id;type:uuid;primaryKey"`
	UserID       string    `gorm:"column:user_id;type:uuid;not null"`
	ProductID    string    `gorm:"column:product_id;type:uuid;not null"`
	PharmacyID   string    `gorm:"column:pharmacy_id;type:uuid;not null"`
	Quantity     int       `gorm:"column:quantity;not null"`
	UnitPrice    float64   `gorm:"column:unit_price;type:decimal(10,2);not null"`
	ProductName  string    `gorm:"column:product_name;type:varchar(255)"`
	PharmacyName string    `gorm:"column:pharmacy_name;type:varchar(255)"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (CartItem) TableName() string {
	return `"orders".cart_items`
}

func (c *CartItem) Subtotal() float64 {
	return c.UnitPrice * float64(c.Quantity)
}
