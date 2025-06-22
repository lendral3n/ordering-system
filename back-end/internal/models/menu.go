package models

import (
	"time"

	"gorm.io/gorm"
)

// MenuCategory model
type MenuCategory struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"not null" json:"name"`
	Description  *string        `json:"description"`
	DisplayOrder int            `gorm:"default:0" json:"display_order"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	MenuItems []MenuItem `json:"menu_items,omitempty"`
}

// MenuItem model
type MenuItem struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CategoryID      uint           `gorm:"not null" json:"category_id"`
	Name            string         `gorm:"not null" json:"name"`
	Description     *string        `json:"description"`
	Price           float64        `gorm:"not null" json:"price"`
	ImageURL        *string        `json:"image_url"`
	Image360URL     *string        `json:"image_360_url"`
	VideoURL        *string        `json:"video_url"`
	IsAvailable     bool           `gorm:"default:true" json:"is_available"`
	PreparationTime *int           `json:"preparation_time"` // in minutes
	StockQuantity   *int           `json:"stock_quantity"`   // NULL = unlimited
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	Category   MenuCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	MediaFiles []MediaFile  `json:"media_files,omitempty"`
}