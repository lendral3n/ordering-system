package repository

import (
	"lendral3n/ordering-system/internal/models"
	"time"
	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) GetByOrderID(orderID uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("order_id = ?", orderID).
		Order("created_at DESC").
		First(&payment).Error
	
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByMidtransOrderID(midtransOrderID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("midtrans_order_id = ?", midtransOrderID).
		First(&payment).Error
	
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) Update(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

func (r *paymentRepository) GetAll(filter PaymentFilter) ([]models.Payment, error) {
	var payments []models.Payment
	query := r.db.Preload("Order.Table")
	
	// Apply filters
	if filter.Status != nil {
		query = query.Where("transaction_status = ?", *filter.Status)
	}
	
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", filter.DateTo.Add(24*time.Hour))
	}
	
	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	
	err := query.Order("created_at DESC").Find(&payments).Error
	return payments, err
}