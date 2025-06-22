package models

import (
	"time"

	"gorm.io/gorm"
)

// MediaFile model
type MediaFile struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	FileType     string         `gorm:"not null" json:"file_type"` // image, video, image_360
	FileURL      string         `gorm:"not null" json:"file_url"`
	PublicID     string         `json:"public_id"` // Cloudinary public ID
	ThumbnailURL *string        `json:"thumbnail_url"`
	FileSize     int64          `json:"file_size"`
	MimeType     string         `json:"mime_type"`
	MenuItemID   *uint          `json:"menu_item_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	MenuItem *MenuItem `gorm:"foreignKey:MenuItemID" json:"menu_item,omitempty"`
}

// Media file type constants
const (
	MediaFileTypeImage    = "image"
	MediaFileTypeVideo    = "video"
	MediaFileTypeImage360 = "image_360"
)