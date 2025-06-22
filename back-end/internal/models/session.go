
// internal/models/session.go
package models

import (
	"time"
)

type CustomerSession struct {
	ID            int        `json:"id" db:"id"`
	SessionToken  string     `json:"session_token" db:"session_token"`
	TableID       int        `json:"table_id" db:"table_id"`
	CustomerName  *string    `json:"customer_name" db:"customer_name"`
	CustomerPhone *string    `json:"customer_phone" db:"customer_phone"`
	StartedAt     time.Time  `json:"started_at" db:"started_at"`
	EndedAt       *time.Time `json:"ended_at" db:"ended_at"`
	
	// Relations
	Table  *Table  `json:"table,omitempty"`
	Orders []Order `json:"orders,omitempty"`
}