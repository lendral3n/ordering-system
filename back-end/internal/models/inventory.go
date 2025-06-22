package models

import (
	"time"

	"gorm.io/gorm"
)

// InventoryLog model
type InventoryLog struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	MenuItemID     uint           `gorm:"not null" json:"menu_item_id"`
	QuantityChange int            `gorm:"not null" json:"quantity_change"`
	Reason         string         `json:"reason"`
	OrderItemID    *uint          `json:"order_item_id"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	MenuItem  MenuItem   `gorm:"foreignKey:MenuItemID" json:"menu_item,omitempty"`
	OrderItem *OrderItem `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
}