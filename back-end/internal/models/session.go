package models

import (
	"time"

	"gorm.io/gorm"
)

// CustomerSession model
type CustomerSession struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	SessionToken  string         `gorm:"uniqueIndex;not null" json:"session_token"`
	TableID       uint           `gorm:"not null" json:"table_id"`
	CustomerName  *string        `json:"customer_name"`
	CustomerPhone *string        `json:"customer_phone"`
	StartedAt     time.Time      `gorm:"autoCreateTime" json:"started_at"`
	EndedAt       *time.Time     `json:"ended_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Table  Table   `gorm:"foreignKey:TableID" json:"table,omitempty"`
	Orders []Order `gorm:"foreignKey:SessionID;references:ID" json:"orders,omitempty"`
}
