// internal/models/inventory.go
package models

import (
	"time"
)

type InventoryLog struct {
	ID             int       `json:"id" db:"id"`
	MenuItemID     int       `json:"menu_item_id" db:"menu_item_id"`
	QuantityChange int       `json:"quantity_change" db:"quantity_change"`
	Reason         string    `json:"reason" db:"reason"`
	OrderItemID    *int      `json:"order_item_id" db:"order_item_id"`
	StaffID        *int      `json:"staff_id" db:"staff_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`

	// Relations
	MenuItem  *MenuItem  `json:"menu_item,omitempty"`
	OrderItem *OrderItem `json:"order_item,omitempty"`
	Staff     *Staff     `json:"staff,omitempty"`
}
