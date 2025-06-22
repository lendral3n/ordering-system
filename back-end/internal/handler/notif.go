package handlers

import (
	"lendral3n/ordering-system/internal/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func (h *Handlers) GetNotifications(c *fiber.Ctx) error {
	unreadOnly := c.Query("unread_only") == "true"

	var notifications []models.Notification
	query := h.DB.Preload("Order.Table")

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	if err := query.Order("created_at DESC").Limit(50).Find(&notifications).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get notifications",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Notifications retrieved",
		"data":    notifications,
	})
}

func (h *Handlers) GetUnreadCount(c *fiber.Ctx) error {
	var count int64
	if err := h.DB.Model(&models.Notification{}).Where("is_read = ?", false).Count(&count).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get unread count",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Unread count retrieved",
		"data": map[string]int64{
			"unread_count": count,
		},
	})
}

func (h *Handlers) MarkAsRead(c *fiber.Ctx) error {
	notifID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid notification ID",
		})
	}

	result := h.DB.Model(&models.Notification{}).
		Where("id = ?", notifID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to mark as read",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Notification not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Notification marked as read",
	})
}

func (h *Handlers) MarkAllAsRead(c *fiber.Ctx) error {
	if err := h.DB.Model(&models.Notification{}).
		Where("is_read = ?", false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to mark all as read",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "All notifications marked as read",
	})
}