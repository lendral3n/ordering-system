// internal/models/payment.go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Payment struct {
	ID                    int             `json:"id" db:"id"`
	OrderID               int             `json:"order_id" db:"order_id"`
	MidtransOrderID       string          `json:"midtrans_order_id" db:"midtrans_order_id"`
	MidtransTransactionID *string         `json:"midtrans_transaction_id" db:"midtrans_transaction_id"`
	PaymentType           *string         `json:"payment_type" db:"payment_type"`
	TransactionStatus     *string         `json:"transaction_status" db:"transaction_status"`
	TransactionTime       *time.Time      `json:"transaction_time" db:"transaction_time"`
	GrossAmount           float64         `json:"gross_amount" db:"gross_amount"`
	Currency              string          `json:"currency" db:"currency"`
	VANumber              *string         `json:"va_number" db:"va_number"`
	Bank                  *string         `json:"bank" db:"bank"`
	FraudStatus           *string         `json:"fraud_status" db:"fraud_status"`
	SignatureKey          *string         `json:"signature_key" db:"signature_key"`
	StatusMessage         *string         `json:"status_message" db:"status_message"`
	RawResponse           json.RawMessage `json:"raw_response" db:"raw_response"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at" db:"updated_at"`

	// Relations
	Order *Order `json:"order,omitempty"`
}


// Define a local type that wraps json.RawMessage
type JSONRawMessage json.RawMessage

// Implement sql.Scanner and driver.Valuer for JSONRawMessage
func (j *JSONRawMessage) Scan(value interface{}) error {
	if value == nil {
		*j = JSONRawMessage("null")
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONRawMessage", value)
	}
	*j = JSONRawMessage(bytes)
	return nil
}

func (j JSONRawMessage) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}
