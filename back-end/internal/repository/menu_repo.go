
// internal/repository/menu_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/models"
)

type menuRepository struct {
	db *database.DB
}

func NewMenuRepository(db *database.DB) MenuRepository {
	return &menuRepository{db: db}
}

func (r *menuRepository) GetCategories(activeOnly bool) ([]models.MenuCategory, error) {
	query := `
		SELECT id, name, description, display_order, is_active, created_at, updated_at
		FROM menu_categories
	`
	if activeOnly {
		query += " WHERE is_active = true"
	}
	query += " ORDER BY display_order, name"
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var categories []models.MenuCategory
	for rows.Next() {
		var cat models.MenuCategory
		err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.Description,
			&cat.DisplayOrder,
			&cat.IsActive,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	
	return categories, rows.Err()
}

func (r *menuRepository) GetCategoryByID(id int) (*models.MenuCategory, error) {
	query := `
		SELECT id, name, description, display_order, is_active, created_at, updated_at
		FROM menu_categories
		WHERE id = $1
	`
	
	var cat models.MenuCategory
	err := r.db.QueryRow(query, id).Scan(
		&cat.ID,
		&cat.Name,
		&cat.Description,
		&cat.DisplayOrder,
		&cat.IsActive,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category not found")
	}
	
	return &cat, err
}

func (r *menuRepository) GetMenuItems(categoryID *int, availableOnly bool) ([]models.MenuItem, error) {
	query := `
		SELECT 
			mi.id, mi.category_id, mi.name, mi.description, mi.price,
			mi.image_url, mi.image_360_url, mi.video_url, mi.is_available,
			mi.preparation_time, mi.stock_quantity, mi.created_at, mi.updated_at,
			mc.id, mc.name, mc.description, mc.display_order, mc.is_active,
			mc.created_at, mc.updated_at
		FROM menu_items mi
		LEFT JOIN menu_categories mc ON mi.category_id = mc.id
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 0
	
	if categoryID != nil {
		argCount++
		query += fmt.Sprintf(" AND mi.category_id = $%d", argCount)
		args = append(args, *categoryID)
	}
	
	if availableOnly {
		query += " AND mi.is_available = true"
	}
	
	query += " ORDER BY mc.display_order, mi.name"
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var items []models.MenuItem
	for rows.Next() {
		var item models.MenuItem
		var cat models.MenuCategory
		
		err := rows.Scan(
			&item.ID,
			&item.CategoryID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.ImageURL,
			&item.Image360URL,
			&item.VideoURL,
			&item.IsAvailable,
			&item.PreparationTime,
			&item.StockQuantity,
			&item.CreatedAt,
			&item.UpdatedAt,
			&cat.ID,
			&cat.Name,
			&cat.Description,
			&cat.DisplayOrder,
			&cat.IsActive,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		item.Category = &cat
		items = append(items, item)
	}
	
	return items, rows.Err()
}

func (r *menuRepository) GetMenuItemByID(id int) (*models.MenuItem, error) {
	query := `
		SELECT 
			id, category_id, name, description, price,
			image_url, image_360_url, video_url, is_available,
			preparation_time, stock_quantity, created_at, updated_at
		FROM menu_items
		WHERE id = $1
	`
	
	var item models.MenuItem
	err := r.db.QueryRow(query, id).Scan(
		&item.ID,
		&item.CategoryID,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.ImageURL,
		&item.Image360URL,
		&item.VideoURL,
		&item.IsAvailable,
		&item.PreparationTime,
		&item.StockQuantity,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("menu item not found")
	}
	
	return &item, err
}

func (r *menuRepository) CreateCategory(category *models.MenuCategory) error {
	query := `
		INSERT INTO menu_categories (name, description, display_order, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	
	return r.db.QueryRow(
		query,
		category.Name,
		category.Description,
		category.DisplayOrder,
		category.IsActive,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

func (r *menuRepository) UpdateCategory(category *models.MenuCategory) error {
	query := `
		UPDATE menu_categories
		SET name = $2, description = $3, display_order = $4, is_active = $5
		WHERE id = $1
		RETURNING updated_at
	`
	
	return r.db.QueryRow(
		query,
		category.ID,
		category.Name,
		category.Description,
		category.DisplayOrder,
		category.IsActive,
	).Scan(&category.UpdatedAt)
}

func (r *menuRepository) CreateMenuItem(item *models.MenuItem) error {
	query := `
		INSERT INTO menu_items (
			category_id, name, description, price, image_url,
			image_360_url, video_url, is_available, preparation_time, stock_quantity
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	
	return r.db.QueryRow(
		query,
		item.CategoryID,
		item.Name,
		item.Description,
		item.Price,
		item.ImageURL,
		item.Image360URL,
		item.VideoURL,
		item.IsAvailable,
		item.PreparationTime,
		item.StockQuantity,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (r *menuRepository) UpdateMenuItem(item *models.MenuItem) error {
	query := `
		UPDATE menu_items
		SET category_id = $2, name = $3, description = $4, price = $5,
			image_url = $6, image_360_url = $7, video_url = $8,
			is_available = $9, preparation_time = $10, stock_quantity = $11
		WHERE id = $1
		RETURNING updated_at
	`
	
	return r.db.QueryRow(
		query,
		item.ID,
		item.CategoryID,
		item.Name,
		item.Description,
		item.Price,
		item.ImageURL,
		item.Image360URL,
		item.VideoURL,
		item.IsAvailable,
		item.PreparationTime,
		item.StockQuantity,
	).Scan(&item.UpdatedAt)
}

func (r *menuRepository) UpdateStock(id int, quantity int) error {
	query := `
		UPDATE menu_items
		SET stock_quantity = stock_quantity + $2
		WHERE id = $1 AND stock_quantity IS NOT NULL
	`
	
	result, err := r.db.Exec(query, id, quantity)
	if err != nil {
		return err
	}
	
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if affected == 0 {
		return fmt.Errorf("menu item not found or stock not tracked")
	}
	
	return nil
}

func (r *menuRepository) DeleteMenuItem(id int) error {
	query := `DELETE FROM menu_items WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}