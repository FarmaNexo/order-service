package entities

import "time"

type OrderStatusHistory struct {
	ID              string    `gorm:"column:id;type:uuid;primaryKey"`
	OrderID         string    `gorm:"column:order_id;type:uuid;not null"`
	Status          string    `gorm:"column:status;type:varchar(20);not null"`
	Notes           *string   `gorm:"column:notes;type:text"`
	ChangedByUserID *string   `gorm:"column:changed_by_user_id;type:uuid"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (OrderStatusHistory) TableName() string {
	return `"orders".order_status_history`
}
