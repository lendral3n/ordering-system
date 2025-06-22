// internal/repository/notification_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/models"
)

type notificationRepository struct {
	db *database.DB
}

func NewNotificationRepository(db *database.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *models.StaffNotification) error {
	query := `
		INSERT INTO staff_notifications (staff_id, order_id, type, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		query,
		notification.StaffID,
		notification.OrderID,
		notification.Type,
		notification.Message,
	).Scan(&notification.ID, &notification.CreatedAt)
}

func (r *notificationRepository) GetByStaffID(staffID int, unreadOnly bool) ([]models.StaffNotification, error) {
	query := `
		SELECT 
			n.id, n.staff_id, n.order_id, n.type, n.message,
			n.is_read, n.read_at, n.created_at,
			o.id, o.order_number, o.table_id, o.status, o.grand_total,
			o.payment_status, o.created_at
		FROM staff_notifications n
		LEFT JOIN orders o ON n.order_id = o.id
		WHERE n.staff_id = $1
	`

	if unreadOnly {
		query += " AND n.is_read = false"
	}

	query += " ORDER BY n.created_at DESC LIMIT 50"

	rows, err := r.db.Query(query, staffID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.StaffNotification
	for rows.Next() {
		var notif models.StaffNotification
		var order models.Order
		var orderID, tableID sql.NullInt64
		var orderNumber, status, paymentStatus sql.NullString
		var grandTotal sql.NullFloat64
		var orderCreatedAt sql.NullTime

		err := rows.Scan(
			&notif.ID,
			&notif.StaffID,
			&notif.OrderID,
			&notif.Type,
			&notif.Message,
			&notif.IsRead,
			&notif.ReadAt,
			&notif.CreatedAt,
			&orderID,
			&orderNumber,
			&tableID,
			&status,
			&grandTotal,
			&paymentStatus,
			&orderCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Set order if exists
		if orderID.Valid {
			order.ID = int(orderID.Int64)
			order.OrderNumber = orderNumber.String
			order.TableID = int(tableID.Int64)
			order.Status = models.OrderStatus(status.String)
			order.GrandTotal = grandTotal.Float64
			order.PaymentStatus = models.PaymentStatus(paymentStatus.String)
			order.CreatedAt = orderCreatedAt.Time
			notif.Order = &order
		}

		notifications = append(notifications, notif)
	}

	return notifications, rows.Err()
}

func (r *notificationRepository) MarkAsRead(id int) error {
	query := `
		UPDATE staff_notifications
		SET is_read = true, read_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (r *notificationRepository) MarkAllAsRead(staffID int) error {
	query := `
		UPDATE staff_notifications
		SET is_read = true, read_at = CURRENT_TIMESTAMP
		WHERE staff_id = $1 AND is_read = false
	`

	_, err := r.db.Exec(query, staffID)
	return err
}

func (r *notificationRepository) GetUnreadCount(staffID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM staff_notifications
		WHERE staff_id = $1 AND is_read = false
	`

	var count int
	err := r.db.QueryRow(query, staffID).Scan(&count)
	return count, err
}
