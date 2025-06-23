package repository

import (
	"fmt"
	"lendral3n/ordering-system/internal/models"
	"time"
	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *models.Order) error {
	// Generate order number if not exists
	if order.OrderNumber == "" {
		order.OrderNumber = r.generateOrderNumber()
	}
	
	return r.db.Create(order).Error
}

func (r *orderRepository) GetByID(id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Table").
		Preload("OrderItems.MenuItem").
		Preload("Payment").
		First(&order, id).Error
	
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNumber(orderNumber string) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Table").
		Preload("OrderItems.MenuItem").
		Where("order_number = ?", orderNumber).
		First(&order).Error
	
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByTableID(tableID uint, status []string) ([]models.Order, error) {
	var orders []models.Order
	query := r.db.Preload("OrderItems").Where("table_id = ?", tableID)
	
	if len(status) > 0 {
		query = query.Where("status IN ?", status)
	}
	
	err := query.Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetBySessionID(sessionID uint) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("OrderItems.MenuItem").
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Find(&orders).Error
	
	return orders, err
}

func (r *orderRepository) GetAll(filter OrderFilter) ([]models.Order, error) {
	var orders []models.Order
	query := r.db.Preload("Table").Preload("OrderItems.MenuItem")
	
	// Apply filters
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.PaymentStatus != nil {
		query = query.Where("payment_status = ?", *filter.PaymentStatus)
	}
	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", *filter.DateTo)
	}
	
	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	
	err := query.Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error
}

func (r *orderRepository) UpdatePaymentStatus(id uint, status string, method *string) error {
	updates := map[string]interface{}{
		"payment_status": status,
	}
	if method != nil {
		updates["payment_method"] = *method
	}
	
	return r.db.Model(&models.Order{}).Where("id = ?", id).Updates(updates).Error
}

func (r *orderRepository) GetOrderItems(orderID uint) ([]models.OrderItem, error) {
	var items []models.OrderItem
	err := r.db.Preload("MenuItem").Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

func (r *orderRepository) UpdateOrderItemStatus(id uint, status string) error {
	return r.db.Model(&models.OrderItem{}).Where("id = ?", id).Update("status", status).Error
}

func (r *orderRepository) generateOrderNumber() string {
	// Format: ORD-YYYYMMDD-XXXXX
	now := time.Now()
	dateStr := now.Format("20060102")
	
	// Get today's order count
	var count int64
	r.db.Model(&models.Order{}).
		Where("DATE(created_at) = DATE(?)", now).
		Count(&count)
	
	return fmt.Sprintf("ORD-%s-%05d", dateStr, count+1)
}