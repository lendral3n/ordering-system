package models

import (
	"time"

	"gorm.io/gorm"
)

// Payment model
type Payment struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	OrderID               uint           `gorm:"not null" json:"order_id"`
	MidtransOrderID       string         `gorm:"uniqueIndex" json:"midtrans_order_id"`
	MidtransTransactionID *string        `json:"midtrans_transaction_id"`
	PaymentType           *string        `json:"payment_type"`
	TransactionStatus     *string        `json:"transaction_status"`
	TransactionTime       *time.Time     `json:"transaction_time"`
	GrossAmount           float64        `gorm:"not null" json:"gross_amount"`
	Currency              string         `gorm:"default:'IDR'" json:"currency"`
	VANumber              *string        `json:"va_number"`
	Bank                  *string        `json:"bank"`
	FraudStatus           *string        `json:"fraud_status"`
	StatusMessage         *string        `json:"status_message"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	Order Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}