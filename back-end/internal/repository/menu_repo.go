package repository

import (
	"lendral3n/ordering-system/internal/models"
	"gorm.io/gorm"
)

type menuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepository{db: db}
}

func (r *menuRepository) GetCategories(activeOnly bool) ([]models.MenuCategory, error) {
	var categories []models.MenuCategory
	query := r.db.Order("display_order, name")
	
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	
	err := query.Find(&categories).Error
	return categories, err
}

func (r *menuRepository) GetCategoryByID(id uint) (*models.MenuCategory, error) {
	var category models.MenuCategory
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *menuRepository) GetMenuItems(categoryID *uint, availableOnly bool) ([]models.MenuItem, error) {
	var items []models.MenuItem
	query := r.db.Preload("Category").Preload("MediaFiles")
	
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	
	if availableOnly {
		query = query.Where("is_available = ?", true)
	}
	
	err := query.Order("name").Find(&items).Error
	return items, err
}

func (r *menuRepository) GetMenuItemByID(id uint) (*models.MenuItem, error) {
	var item models.MenuItem
	err := r.db.Preload("Category").Preload("MediaFiles").First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *menuRepository) CreateCategory(category *models.MenuCategory) error {
	return r.db.Create(category).Error
}

func (r *menuRepository) UpdateCategory(category *models.MenuCategory) error {
	return r.db.Save(category).Error
}

func (r *menuRepository) CreateMenuItem(item *models.MenuItem) error {
	return r.db.Create(item).Error
}

func (r *menuRepository) UpdateMenuItem(item *models.MenuItem) error {
	return r.db.Save(item).Error
}

func (r *menuRepository) UpdateStock(id uint, quantity int) error {
	result := r.db.Model(&models.MenuItem{}).
		Where("id = ? AND stock_quantity IS NOT NULL", id).
		Update("stock_quantity", gorm.Expr("stock_quantity + ?", quantity))
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	
	return nil
}

func (r *menuRepository) DeleteMenuItem(id uint) error {
	return r.db.Delete(&models.MenuItem{}, id).Error
}