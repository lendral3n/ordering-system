package repository

import (
	"lendral3n/ordering-system/internal/models"
	"time"
)

// Repository interfaces for clean architecture
type Repositories struct {
	Table        TableRepository
	Menu         MenuRepository
	Order        OrderRepository
	Payment      PaymentRepository
	Session      SessionRepository
	Notification NotificationRepository
	Media        MediaRepository
	Staff        StaffRepository
}

type StaffRepository interface {
	GetByID(id uint) (*models.Staff, error)
	GetByUsername(username string) (*models.Staff, error)
	GetByEmail(email string) (*models.Staff, error)
	Create(staff *models.Staff) error
	Update(staff *models.Staff) error
	UpdateLastLogin(id uint) error
	GetAll(activeOnly bool) ([]models.Staff, error)
}

type TableRepository interface {
	GetByID(id uint) (*models.Table, error)
	GetByTableNumber(tableNumber string) (*models.Table, error)
	GetAll() ([]models.Table, error)
	Create(table *models.Table) error
	Update(table *models.Table) error
	UpdateStatus(id uint, status string) error
}

type MenuRepository interface {
	GetCategories(activeOnly bool) ([]models.MenuCategory, error)
	GetCategoryByID(id uint) (*models.MenuCategory, error)
	GetMenuItems(categoryID *uint, availableOnly bool) ([]models.MenuItem, error)
	GetMenuItemByID(id uint) (*models.MenuItem, error)
	CreateCategory(category *models.MenuCategory) error
	UpdateCategory(category *models.MenuCategory) error
	CreateMenuItem(item *models.MenuItem) error
	UpdateMenuItem(item *models.MenuItem) error
	UpdateStock(id uint, quantity int) error
	DeleteMenuItem(id uint) error
}

type OrderRepository interface {
	Create(order *models.Order) error
	GetByID(id uint) (*models.Order, error)
	GetByOrderNumber(orderNumber string) (*models.Order, error)
	GetByTableID(tableID uint, status []string) ([]models.Order, error)
	GetBySessionID(sessionID uint) ([]models.Order, error)
	GetAll(filter OrderFilter) ([]models.Order, error)
	UpdateStatus(id uint, status string) error
	UpdatePaymentStatus(id uint, status string, method *string) error
	GetOrderItems(orderID uint) ([]models.OrderItem, error)
	UpdateOrderItemStatus(id uint, status string) error
}

type PaymentRepository interface {
	Create(payment *models.Payment) error
	GetByOrderID(orderID uint) (*models.Payment, error)
	GetByMidtransOrderID(midtransOrderID string) (*models.Payment, error)
	Update(payment *models.Payment) error
	GetAll(filter PaymentFilter) ([]models.Payment, error)
}

type SessionRepository interface {
	Create(session *models.CustomerSession) error
	GetByToken(token string) (*models.CustomerSession, error)
	GetByTableID(tableID uint, activeOnly bool) (*models.CustomerSession, error)
	EndSession(token string) error
	CleanupExpiredSessions(expiryDuration time.Duration) error
}

type NotificationRepository interface {
	Create(notification *models.Notification) error
	GetUnreadOnly(unreadOnly bool) ([]models.Notification, error)
	MarkAsRead(id uint) error
	MarkAllAsRead() error
	GetUnreadCount() (int64, error)
}

type MediaRepository interface {
	Create(media *models.MediaFile) error
	GetByMenuItemID(menuItemID uint) ([]models.MediaFile, error)
	Delete(id uint) error
}

// Filter structs
type OrderFilter struct {
	Status        *string
	PaymentStatus *string
	TableID       *uint
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