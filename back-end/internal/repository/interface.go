// internal/repository/interfaces.go
package repository

import (
	"lendral3n/ordering-system/internal/models"
	"time"
)

type TableRepository interface {
	GetByID(id int) (*models.Table, error)
	GetByTableNumber(tableNumber string) (*models.Table, error)
	GetAll() ([]models.Table, error)
	Create(table *models.Table) error
	Update(table *models.Table) error
	UpdateStatus(id int, status models.TableStatus) error
}

type MenuRepository interface {
	GetCategories(activeOnly bool) ([]models.MenuCategory, error)
	GetCategoryByID(id int) (*models.MenuCategory, error)
	GetMenuItems(categoryID *int, availableOnly bool) ([]models.MenuItem, error)
	GetMenuItemByID(id int) (*models.MenuItem, error)
	CreateCategory(category *models.MenuCategory) error
	UpdateCategory(category *models.MenuCategory) error
	CreateMenuItem(item *models.MenuItem) error
	UpdateMenuItem(item *models.MenuItem) error
	UpdateStock(id int, quantity int) error
	DeleteMenuItem(id int) error
}

type OrderRepository interface {
	Create(order *models.Order) (*models.Order, error)
	GetByID(id int) (*models.Order, error)
	GetByOrderNumber(orderNumber string) (*models.Order, error)
	GetByTableID(tableID int, status []models.OrderStatus) ([]models.Order, error)
	GetBySessionID(sessionID int) ([]models.Order, error)
	GetAll(filter OrderFilter) ([]models.Order, error)
	UpdateStatus(id int, status models.OrderStatus) error
	UpdatePaymentStatus(id int, status models.PaymentStatus, method *string) error
	GetOrderItems(orderID int) ([]models.OrderItem, error)
	UpdateOrderItemStatus(id int, status models.OrderItemStatus) error
}

type PaymentRepository interface {
	Create(payment *models.Payment) error
	GetByOrderID(orderID int) (*models.Payment, error)
	GetByMidtransOrderID(midtransOrderID string) (*models.Payment, error)
	Update(payment *models.Payment) error
	GetAll(filter PaymentFilter) ([]models.Payment, error)
}

type StaffRepository interface {
	GetByID(id int) (*models.Staff, error)
	GetByUsername(username string) (*models.Staff, error)
	GetByEmail(email string) (*models.Staff, error)
	Create(staff *models.Staff) error
	Update(staff *models.Staff) error
	UpdateLastLogin(id int) error
	GetAll(activeOnly bool) ([]models.Staff, error)
}

type SessionRepository interface {
	Create(session *models.CustomerSession) error
	GetByToken(token string) (*models.CustomerSession, error)
	GetByTableID(tableID int, activeOnly bool) (*models.CustomerSession, error)
	EndSession(token string) error
	CleanupExpiredSessions(expiryDuration time.Duration) error
}

type NotificationRepository interface {
	Create(notification *models.StaffNotification) error
	GetByStaffID(staffID int, unreadOnly bool) ([]models.StaffNotification, error)
	MarkAsRead(id int) error
	MarkAllAsRead(staffID int) error
	GetUnreadCount(staffID int) (int, error)
}

type MediaRepository interface {
	Create(media *models.MediaFile) error
	GetByMenuItemID(menuItemID int) ([]models.MediaFile, error)
	Delete(id int) error
}

// Filter structs
type OrderFilter struct {
	Status        *models.OrderStatus
	PaymentStatus *models.PaymentStatus
	TableID       *int
	DateFrom      *time.Time
	DateTo        *time.Time
	Limit         int
	Offset        int
}

type PaymentFilter struct {
	Status   *string
	DateFrom *time.Time
	DateTo   *time.Time
	Limit    int
	Offset   int
}
