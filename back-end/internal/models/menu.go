
// internal/models/menu.go
package models

import (
	"database/sql"
	"time"
)

type MenuCategory struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  *string   `json:"description" db:"description"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	
	// Relations
	MenuItems []MenuItem `json:"menu_items,omitempty"`
}

type MenuItem struct {
	ID              int            `json:"id" db:"id"`
	CategoryID      int            `json:"category_id" db:"category_id"`
	Name            string         `json:"name" db:"name"`
	Description     *string        `json:"description" db:"description"`
	Price           float64        `json:"price" db:"price"`
	ImageURL        *string        `json:"image_url" db:"image_url"`
	Image360URL     *string        `json:"image_360_url" db:"image_360_url"`
	VideoURL        *string        `json:"video_url" db:"video_url"`
	IsAvailable     bool           `json:"is_available" db:"is_available"`
	PreparationTime sql.NullInt64  `json:"preparation_time" db:"preparation_time"`
	StockQuantity   sql.NullInt64  `json:"stock_quantity" db:"stock_quantity"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	
	// Relations
	Category    *MenuCategory `json:"category,omitempty"`
	MediaFiles  []MediaFile   `json:"media_files,omitempty"`
}
