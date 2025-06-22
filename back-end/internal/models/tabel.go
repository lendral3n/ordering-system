// internal/models/table.go
package models

import (
	"time"
)

type TableStatus string

const (
	TableStatusAvailable TableStatus = "available"
	TableStatusOccupied  TableStatus = "occupied"
	TableStatusReserved  TableStatus = "reserved"
)

type Table struct {
	ID          int         `json:"id" db:"id"`
	TableNumber string      `json:"table_number" db:"table_number"`
	QRCode      string      `json:"qr_code" db:"qr_code"`
	Status      TableStatus `json:"status" db:"status"`
	Capacity    int         `json:"capacity" db:"capacity"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}





