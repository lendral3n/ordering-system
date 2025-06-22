package models

import (
	"time"

	"gorm.io/gorm"
)

// Notification model
type Notification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	OrderID   *uint          `json:"order_id"`
	Type      string         `gorm:"not null" json:"type"` // new_order, payment_received, order_ready, assistance_request
	Message   string         `gorm:"not null" json:"message"`
	IsRead    bool           `gorm:"default:false" json:"is_read"`
	ReadAt    *time.Time     `json:"read_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// Notification type constants
const (
	NotificationNewOrder          = "new_order"
	NotificationPaymentReceived   = "payment_received"
	NotificationOrderReady        = "order_ready"
	NotificationAssistanceRequest = "assistance_request"
)