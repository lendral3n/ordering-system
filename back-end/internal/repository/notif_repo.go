package repository

import (
	"lendral3n/ordering-system/internal/models"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetUnreadOnly(unreadOnly bool) ([]models.Notification, error) {
	var notifications []models.Notification
	query := r.db.Preload("Order.Table")
	
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}
	
	err := query.Order("created_at DESC").Limit(50).Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) MarkAsRead(id uint) error {
	result := r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	
	return nil
}

func (r *notificationRepository) MarkAllAsRead() error {
	return r.db.Model(&models.Notification{}).
		Where("is_read = ?", false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

func (r *notificationRepository) GetUnreadCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("is_read = ?", false).
		Count(&count).Error
	
	return count, err
}