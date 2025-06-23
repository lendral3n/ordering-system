package repository

import (
	"lendral3n/ordering-system/internal/models"
	"gorm.io/gorm"
)

type tableRepository struct {
	db *gorm.DB
}

func NewTableRepository(db *gorm.DB) TableRepository {
	return &tableRepository{db: db}
}

func (r *tableRepository) GetByID(id uint) (*models.Table, error) {
	var table models.Table
	if err := r.db.First(&table, id).Error; err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *tableRepository) GetByTableNumber(tableNumber string) (*models.Table, error) {
	var table models.Table
	if err := r.db.Where("table_number = ?", tableNumber).First(&table).Error; err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *tableRepository) GetAll() ([]models.Table, error) {
	var tables []models.Table
	if err := r.db.Order("table_number").Find(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

func (r *tableRepository) Create(table *models.Table) error {
	return r.db.Create(table).Error
}

func (r *tableRepository) Update(table *models.Table) error {
	return r.db.Save(table).Error
}

func (r *tableRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Table{}).Where("id = ?", id).Update("status", status).Error
}