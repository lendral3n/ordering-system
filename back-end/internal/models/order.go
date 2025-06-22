
// internal/models/order.go
package models

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusReady     OrderStatus = "ready"
	OrderStatusServed    OrderStatus = "served"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid   PaymentStatus = "unpaid"
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type Order struct {
	ID            int           `json:"id" db:"id"`
	OrderNumber   string        `json:"order_number" db:"order_number"`
	SessionID     int           `json:"session_id" db:"session_id"`
	TableID       int           `json:"table_id" db:"table_id"`
	Status        OrderStatus   `json:"status" db:"status"`
	TotalAmount   float64       `json:"total_amount" db:"total_amount"`
	TaxAmount     float64       `json:"tax_amount" db:"tax_amount"`
	ServiceCharge float64       `json:"service_charge" db:"service_charge"`
	GrandTotal    float64       `json:"grand_total" db:"grand_total"`
	PaymentStatus PaymentStatus `json:"payment_status" db:"payment_status"`
	PaymentMethod *string       `json:"payment_method" db:"payment_method"`
	Notes         *string       `json:"notes" db:"notes"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	
	// Relations
	OrderItems      []OrderItem      `json:"order_items,omitempty"`
	Table           *Table           `json:"table,omitempty"`
	Payment         *Payment         `json:"payment,omitempty"`
	CustomerSession *CustomerSession `json:"customer_session,omitempty"`
}

type OrderItemStatus string

const (
	OrderItemStatusPending   OrderItemStatus = "pending"
	OrderItemStatusPreparing OrderItemStatus = "preparing"
	OrderItemStatusReady     OrderItemStatus = "ready"
	OrderItemStatusServed    OrderItemStatus = "served"
	OrderItemStatusCancelled OrderItemStatus = "cancelled"
)

type OrderItem struct {
	ID         int             `json:"id" db:"id"`
	OrderID    int             `json:"order_id" db:"order_id"`
	MenuItemID int             `json:"menu_item_id" db:"menu_item_id"`
	Quantity   int             `json:"quantity" db:"quantity"`
	UnitPrice  float64         `json:"unit_price" db:"unit_price"`
	Subtotal   float64         `json:"subtotal" db:"subtotal"`
	Notes      *string         `json:"notes" db:"notes"`
	Status     OrderItemStatus `json:"status" db:"status"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
	
	// Relations
	MenuItem *MenuItem `json:"menu_item,omitempty"`
}