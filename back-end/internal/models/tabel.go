package models

import (
	"time"

	"gorm.io/gorm"
)

// Table model
type Table struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TableNumber string         `gorm:"uniqueIndex;not null" json:"table_number"`
	QRCode      string         `gorm:"not null" json:"qr_code"`
	Status      string         `gorm:"default:'available'" json:"status"` // available, occupied, reserved
	Capacity    int            `gorm:"not null" json:"capacity"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Table status constants
const (
	TableStatusAvailable = "available"
	TableStatusOccupied  = "occupied"
	TableStatusReserved  = "reserved"
)