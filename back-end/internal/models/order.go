package models

import (
	"time"

	"gorm.io/gorm"
)

// Order model
type Order struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	OrderNumber   string         `gorm:"uniqueIndex;not null" json:"order_number"`
	SessionID     uint           `gorm:"not null" json:"session_id"`
	TableID       uint           `gorm:"not null" json:"table_id"`
	Status        string         `gorm:"default:'pending'" json:"status"` // pending, confirmed, preparing, ready, served, completed, cancelled
	TotalAmount   float64        `gorm:"not null" json:"total_amount"`
	TaxAmount     float64        `gorm:"default:0" json:"tax_amount"`
	ServiceCharge float64        `gorm:"default:0" json:"service_charge"`
	GrandTotal    float64        `gorm:"not null" json:"grand_total"`
	PaymentStatus string         `gorm:"default:'unpaid'" json:"payment_status"` // unpaid, pending, paid, failed, refunded
	PaymentMethod *string        `json:"payment_method"`
	Notes         *string        `json:"notes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	Session         CustomerSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Table           Table           `gorm:"foreignKey:TableID" json:"table,omitempty"`
	OrderItems      []OrderItem     `json:"order_items,omitempty"`
	Payment         *Payment        `json:"payment,omitempty"`
}

// OrderItem model
type OrderItem struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	OrderID    uint           `gorm:"not null" json:"order_id"`
	MenuItemID uint           `gorm:"not null" json:"menu_item_id"`
	Quantity   int            `gorm:"not null" json:"quantity"`
	UnitPrice  float64        `gorm:"not null" json:"unit_price"`
	Subtotal   float64        `gorm:"not null" json:"subtotal"`
	Notes      *string        `json:"notes"`
	Status     string         `gorm:"default:'pending'" json:"status"` // pending, preparing, ready, served, cancelled
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	Order    Order    `gorm:"foreignKey:OrderID" json:"-"`
	MenuItem MenuItem `gorm:"foreignKey:MenuItemID" json:"menu_item,omitempty"`
}

// Order status constants
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusPreparing = "preparing"
	OrderStatusReady     = "ready"
	OrderStatusServed    = "served"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
)

// Payment status constants
const (
	PaymentStatusUnpaid   = "unpaid"
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

// Order item status constants
const (
	OrderItemStatusPending   = "pending"
	OrderItemStatusPreparing = "preparing"
	OrderItemStatusReady     = "ready"
	OrderItemStatusServed    = "served"
	OrderItemStatusCancelled = "cancelled"
)