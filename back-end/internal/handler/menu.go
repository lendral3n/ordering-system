package handlers

import (
	"context"
	"lendral3n/ordering-system/internal/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Customer endpoints
func (h *Handlers) GetCategories(c *fiber.Ctx) error {
	var categories []models.MenuCategory
	query := h.DB.Where("is_active = ?", true).Order("display_order, name")
	
	if err := query.Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get categories",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Categories retrieved",
		"data":    categories,
	})
}

func (h *Handlers) GetMenuItems(c *fiber.Ctx) error {
	categoryID := c.Query("category_id")
	
	var items []models.MenuItem
	query := h.DB.Preload("Category").Preload("MediaFiles").Where("is_available = ?", true)
	
	if categoryID != "" {
		id, err := strconv.Atoi(categoryID)
		if err == nil {
			query = query.Where("category_id = ?", id)
		}
	}
	
	if err := query.Order("name").Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get menu items",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Menu items retrieved",
		"data":    items,
	})
}

func (h *Handlers) GetMenuItem(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid menu item ID",
		})
	}

	var item models.MenuItem
	if err := h.DB.Preload("Category").Preload("MediaFiles").First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Menu item not found",
		})
	}

	if !item.IsAvailable {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Menu item is not available",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Menu item retrieved",
		"data":    item,
	})
}

func (h *Handlers) GetMenu360View(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid menu item ID",
		})
	}

	var item models.MenuItem
	if err := h.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Menu item not found",
		})
	}

	if item.Image360URL == nil || *item.Image360URL == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "360° view not available for this item",
		})
	}

	response := map[string]interface{}{
		"menu_item_id": item.ID,
		"name":         item.Name,
		"image_360_url": item.Image360URL,
		"video_url":    item.VideoURL,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "360° view data retrieved",
		"data":    response,
	})
}

// Staff endpoints
func (h *Handlers) CreateCategory(c *fiber.Ctx) error {
	var category models.MenuCategory
	if err := c.BodyParser(&category); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if category.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Category name is required",
		})
	}

	if err := h.DB.Create(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create category",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Category created",
		"data":    category,
	})
}

func (h *Handlers) UpdateCategory(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid category ID",
		})
	}

	var category models.MenuCategory
	if err := c.BodyParser(&category); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	result := h.DB.Model(&models.MenuCategory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":          category.Name,
		"description":   category.Description,
		"display_order": category.DisplayOrder,
		"is_active":     category.IsActive,
	})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update category",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Category not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Category updated",
	})
}

func (h *Handlers) CreateMenuItem(c *fiber.Ctx) error {
	var item models.MenuItem
	if err := c.BodyParser(&item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if item.Name == "" || item.CategoryID == 0 || item.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid menu item data",
		})
	}

	if err := h.DB.Create(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create menu item",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Menu item created",
		"data":    item,
	})
}

func (h *Handlers) UpdateMenuItem(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid item ID",
		})
	}

	var item models.MenuItem
	if err := c.BodyParser(&item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	result := h.DB.Model(&models.MenuItem{}).Where("id = ?", id).Updates(item)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update menu item",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Menu item not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Menu item updated",
	})
}

func (h *Handlers) DeleteMenuItem(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid item ID",
		})
	}

	// Get media files to delete
	var mediaFiles []models.MediaFile
	h.DB.Where("menu_item_id = ?", id).Find(&mediaFiles)

	// Delete menu item (soft delete)
	if err := h.DB.Delete(&models.MenuItem{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete menu item",
		})
	}

	// Delete associated media files from Cloudinary
	ctx := context.Background()
	for _, media := range mediaFiles {
		h.CloudinaryService.DeleteFile(ctx, media.PublicID)
		h.DB.Delete(&media)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Menu item deleted",
	})
}

func (h *Handlers) UploadMedia(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid item ID",
		})
	}

	// Get file type
	fileType := c.FormValue("type", "image")

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "No file uploaded",
		})
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to open file",
		})
	}
	defer src.Close()

	// Upload to Cloudinary
	ctx := context.Background()
	result, err := h.CloudinaryService.UploadFile(ctx, src, file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to upload file",
		})
	}

	// Save media record
	mediaFile := models.MediaFile{
		FileType:     fileType,
		FileURL:      result.SecureURL,
		PublicID:     result.PublicID,
		ThumbnailURL: &result.ThumbnailURL,
		FileSize:     result.FileSize,
		MimeType:     file.Header.Get("Content-Type"),
		MenuItemID:   &[]uint{uint(id)}[0],
	}

	if err := h.DB.Create(&mediaFile).Error; err != nil {
		// Rollback - delete uploaded file
		h.CloudinaryService.DeleteFile(ctx, result.PublicID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to save media record",
		})
	}

	// Update menu item URLs based on file type
	if fileType == "image" {
		h.DB.Model(&models.MenuItem{}).Where("id = ?", id).Update("image_url", result.SecureURL)
	} else if fileType == "image_360" {
		url := h.CloudinaryService.Generate360URL(result.PublicID)
		h.DB.Model(&models.MenuItem{}).Where("id = ?", id).Update("image_360_url", url)
	} else if fileType == "video" {
		h.DB.Model(&models.MenuItem{}).Where("id = ?", id).Update("video_url", result.SecureURL)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "File uploaded successfully",
		"data":    mediaFile,
	})
}

func (h *Handlers) UpdateStock(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid item ID",
		})
	}

	var req struct {
		Quantity int    `json:"quantity"`
		Reason   string `json:"reason"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Update stock in transaction
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		// Update stock quantity
		result := tx.Model(&models.MenuItem{}).
			Where("id = ? AND stock_quantity IS NOT NULL", id).
			Update("stock_quantity", gorm.Expr("stock_quantity + ?", req.Quantity))
		
		if result.RowsAffected == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "Menu item not found or stock not tracked")
		}

		// Create inventory log
		log := models.InventoryLog{
			MenuItemID:     uint(id),
			QuantityChange: req.Quantity,
			Reason:         req.Reason,
		}
		return tx.Create(&log).Error
	})

	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).JSON(fiber.Map{
				"success": false,
				"error":   e.Message,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update stock",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Stock updated",
	})
}