package repository

import (
	"lendral3n/ordering-system/internal/models"
	"gorm.io/gorm"
)

type staffRepository struct {
	db *gorm.DB
}

func NewStaffRepository(db *gorm.DB) StaffRepository {
	return &staffRepository{db: db}
}

func (r *staffRepository) GetByID(id uint) (*models.Staff, error) {
	var staff models.Staff
	err := r.db.First(&staff, id).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

func (r *staffRepository) GetByUsername(username string) (*models.Staff, error) {
	var staff models.Staff
	err := r.db.Where("username = ?", username).First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

func (r *staffRepository) GetByEmail(email string) (*models.Staff, error) {
	var staff models.Staff
	err := r.db.Where("email = ?", email).First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

func (r *staffRepository) Create(staff *models.Staff) error {
	return r.db.Create(staff).Error
}

func (r *staffRepository) Update(staff *models.Staff) error {
	return r.db.Save(staff).Error
}

func (r *staffRepository) UpdateLastLogin(id uint) error {
	return r.db.Model(&models.Staff{}).
		Where("id = ?", id).
		Update("last_login", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *staffRepository) GetAll(activeOnly bool) ([]models.Staff, error) {
	var staffList []models.Staff
	query := r.db.Order("full_name")
	
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	
	err := query.Find(&staffList).Error
	return staffList, err
}