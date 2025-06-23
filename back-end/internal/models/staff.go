// internal/models/staff.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// Staff model
type Staff struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	FullName     string         `gorm:"not null" json:"full_name"`
	Role         string         `gorm:"not null" json:"role"` // admin, cashier, waiter, kitchen
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastLogin    *time.Time     `json:"last_login"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Staff role constants
const (
	StaffRoleAdmin   = "admin"
	StaffRoleCashier = "cashier"
	StaffRoleWaiter  = "waiter"
	StaffRoleKitchen = "kitchen"
)

type StaffNotificationType string

// const (
// 	NotificationNewOrder          StaffNotificationType = "new_order"
// 	NotificationPaymentReceived   StaffNotificationType = "payment_received"
// 	NotificationOrderReady        StaffNotificationType = "order_ready"
// 	NotificationAssistanceRequest StaffNotificationType = "assistance_request"
// )

type StaffNotification struct {
	ID        int                   `json:"id" db:"id"`
	StaffID   *int                  `json:"staff_id" db:"staff_id"`
	OrderID   *int                  `json:"order_id" db:"order_id"`
	Type      StaffNotificationType `json:"type" db:"type"`
	Message   string                `json:"message" db:"message"`
	IsRead    bool                  `json:"is_read" db:"is_read"`
	ReadAt    *time.Time            `json:"read_at" db:"read_at"`
	CreatedAt time.Time             `json:"created_at" db:"created_at"`

	// Relations
	Staff *Staff `json:"staff,omitempty"`
	Order *Order `json:"order,omitempty"`
}
