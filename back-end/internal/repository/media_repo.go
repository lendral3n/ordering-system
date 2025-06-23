package repository

import (
	"lendral3n/ordering-system/internal/models"
	"gorm.io/gorm"
)

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(media *models.MediaFile) error {
	return r.db.Create(media).Error
}

func (r *mediaRepository) GetByMenuItemID(menuItemID uint) ([]models.MediaFile, error) {
	var mediaFiles []models.MediaFile
	err := r.db.Where("menu_item_id = ?", menuItemID).
		Order("created_at DESC").
		Find(&mediaFiles).Error
	
	return mediaFiles, err
}

func (r *mediaRepository) Delete(id uint) error {
	result := r.db.Delete(&models.MediaFile{}, id)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	
	return nil
}