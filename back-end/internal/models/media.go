
// internal/models/media.go
package models

import (
	"time"
)

type MediaFileType string

const (
	MediaFileTypeImage    MediaFileType = "image"
	MediaFileTypeVideo    MediaFileType = "video"
	MediaFileTypeImage360 MediaFileType = "image_360"
)

type MediaFile struct {
	ID           int           `json:"id" db:"id"`
	FileType     MediaFileType `json:"file_type" db:"file_type"`
	FileURL      string        `json:"file_url" db:"file_url"`
	ThumbnailURL *string       `json:"thumbnail_url" db:"thumbnail_url"`
	FileSize     int64         `json:"file_size" db:"file_size"`
	MimeType     string        `json:"mime_type" db:"mime_type"`
	MenuItemID   *int          `json:"menu_item_id" db:"menu_item_id"`
	UploadedBy   int           `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	
	// Relations
	MenuItem *MenuItem `json:"menu_item,omitempty"`
	Staff    *Staff    `json:"staff,omitempty"`
}