
// internal/models/staff.go
package models

import (
	"time"
)

type StaffRole string

const (
	StaffRoleAdmin   StaffRole = "admin"
	StaffRoleCashier StaffRole = "cashier"
	StaffRoleWaiter  StaffRole = "waiter"
	StaffRoleKitchen StaffRole = "kitchen"
)

type Staff struct {
	ID           int        `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FullName     string     `json:"full_name" db:"full_name"`
	Role         StaffRole  `json:"role" db:"role"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastLogin    *time.Time `json:"last_login" db:"last_login"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

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