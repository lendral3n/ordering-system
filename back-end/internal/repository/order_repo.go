// internal/repository/order_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/models"
	"strings"
	"time"
)

type orderRepository struct {
	db *database.DB
}

func NewOrderRepository(db *database.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *models.Order) (*models.Order, error) {
	// Generate order number
	order.OrderNumber = r.generateOrderNumber()
	
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	
	// Insert order
	query := `
		INSERT INTO orders (
			order_number, session_id, table_id, status, total_amount,
			tax_amount, service_charge, grand_total, payment_status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	
	err = tx.QueryRow(
		query,
		order.OrderNumber,
		order.SessionID,
		order.TableID,
		order.Status,
		order.TotalAmount,
		order.TaxAmount,
		order.ServiceCharge,
		order.GrandTotal,
		order.PaymentStatus,
		order.Notes,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	// Insert order items
	for i := range order.OrderItems {
		item := &order.OrderItems[i]
		itemQuery := `
			INSERT INTO order_items (
				order_id, menu_item_id, quantity, unit_price, subtotal, notes, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at
		`
		
		err = tx.QueryRow(
			itemQuery,
			order.ID,
			item.MenuItemID,
			item.Quantity,
			item.UnitPrice,
			item.Subtotal,
			item.Notes,
			models.OrderItemStatusPending,
		).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
		
		if err != nil {
			return nil, err
		}
		
		// Update inventory if tracked
		inventoryQuery := `
			UPDATE menu_items
			SET stock_quantity = stock_quantity - $2
			WHERE id = $1 AND stock_quantity IS NOT NULL AND stock_quantity >= $2
		`
		
		result, err := tx.Exec(inventoryQuery, item.MenuItemID, item.Quantity)
		if err != nil {
			return nil, err
		}
		
		// Check if stock was sufficient
		affected, _ := result.RowsAffected()
		if affected == 0 {
			// Check if item exists and has tracked inventory
			var hasStock sql.NullInt64
			checkQuery := `SELECT stock_quantity FROM menu_items WHERE id = $1`
			err = tx.QueryRow(checkQuery, item.MenuItemID).Scan(&hasStock)
			if err == nil && hasStock.Valid && hasStock.Int64 < int64(item.Quantity) {
				return nil, fmt.Errorf("insufficient stock for menu item %d", item.MenuItemID)
			}
		}
		
		// Log inventory change
		logQuery := `
			INSERT INTO inventory_logs (menu_item_id, quantity_change, reason, order_item_id)
			VALUES ($1, $2, 'order_placed', $3)
		`
		_, err = tx.Exec(logQuery, item.MenuItemID, -item.Quantity, item.ID)
		if err != nil {
			return nil, err
		}
	}
	
	// Update table status
	tableQuery := `UPDATE tables SET status = 'occupied' WHERE id = $1`
	_, err = tx.Exec(tableQuery, order.TableID)
	if err != nil {
		return nil, err
	}
	
	// Create notification for staff
	notifQuery := `
		INSERT INTO staff_notifications (order_id, type, message)
		VALUES ($1, 'new_order', $2)
	`
	notifMessage := fmt.Sprintf("New order #%s from table %d", order.OrderNumber, order.TableID)
	_, err = tx.Exec(notifQuery, order.ID, notifMessage)
	if err != nil {
		return nil, err
	}
	
	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	
	// Fetch complete order with relations
	return r.GetByID(order.ID)
}

func (r *orderRepository) GetByID(id int) (*models.Order, error) {
	// Get order
	orderQuery := `
		SELECT 
			o.id, o.order_number, o.session_id, o.table_id, o.status,
			o.total_amount, o.tax_amount, o.service_charge, o.grand_total,
			o.payment_status, o.payment_method, o.notes, o.created_at, o.updated_at,
			t.id, t.table_number, t.qr_code, t.status, t.capacity, t.created_at, t.updated_at
		FROM orders o
		LEFT JOIN tables t ON o.table_id = t.id
		WHERE o.id = $1
	`
	
	var order models.Order
	var table models.Table
	
	err := r.db.QueryRow(orderQuery, id).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.SessionID,
		&order.TableID,
		&order.Status,
		&order.TotalAmount,
		&order.TaxAmount,
		&order.ServiceCharge,
		&order.GrandTotal,
		&order.PaymentStatus,
		&order.PaymentMethod,
		&order.Notes,
		&order.CreatedAt,
		&order.UpdatedAt,
		&table.ID,
		&table.TableNumber,
		&table.QRCode,
		&table.Status,
		&table.Capacity,
		&table.CreatedAt,
		&table.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, err
	}
	
	order.Table = &table
	
	// Get order items
	items, err := r.GetOrderItems(order.ID)
	if err != nil {
		return nil, err
	}
	order.OrderItems = items
	
	return &order, nil
}

func (r *orderRepository) GetByOrderNumber(orderNumber string) (*models.Order, error) {
	query := `SELECT id FROM orders WHERE order_number = $1`
	var id int
	err := r.db.QueryRow(query, orderNumber).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

func (r *orderRepository) GetByTableID(tableID int, status []models.OrderStatus) ([]models.Order, error) {
	query := `
		SELECT id, order_number, session_id, table_id, status,
			   total_amount, tax_amount, service_charge, grand_total,
			   payment_status, payment_method, notes, created_at, updated_at
		FROM orders
		WHERE table_id = $1
	`
	
	args := []interface{}{tableID}
	
	if len(status) > 0 {
		placeholders := make([]string, len(status))
		for i, s := range status {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			args = append(args, s)
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}
	
	query += " ORDER BY created_at DESC"
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.SessionID,
			&order.TableID,
			&order.Status,
			&order.TotalAmount,
			&order.TaxAmount,
			&order.ServiceCharge,
			&order.GrandTotal,
			&order.PaymentStatus,
			&order.PaymentMethod,
			&order.Notes,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	
	return orders, rows.Err()
}

func (r *orderRepository) GetBySessionID(sessionID int) ([]models.Order, error) {
	query := `
		SELECT id, order_number, session_id, table_id, status,
			   total_amount, tax_amount, service_charge, grand_total,
			   payment_status, payment_method, notes, created_at, updated_at
		FROM orders
		WHERE session_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.SessionID,
			&order.TableID,
			&order.Status,
			&order.TotalAmount,
			&order.TaxAmount,
			&order.ServiceCharge,
			&order.GrandTotal,
			&order.PaymentStatus,
			&order.PaymentMethod,
			&order.Notes,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	
	return orders, rows.Err()
}

func (r *orderRepository) GetAll(filter OrderFilter) ([]models.Order, error) {
	query := `
		SELECT 
			o.id, o.order_number, o.session_id, o.table_id, o.status,
			o.total_amount, o.tax_amount, o.service_charge, o.grand_total,
			o.payment_status, o.payment_method, o.notes, o.created_at, o.updated_at,
			t.id, t.table_number, t.qr_code, t.status, t.capacity, t.created_at, t.updated_at
		FROM orders o
		LEFT JOIN tables t ON o.table_id = t.id
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 0
	
	if filter.Status != nil {
		argCount++
		query += fmt.Sprintf(" AND o.status = $%d", argCount)
		args = append(args, *filter.Status)
	}
	
	if filter.PaymentStatus != nil {
		argCount++
		query += fmt.Sprintf(" AND o.payment_status = $%d", argCount)
		args = append(args, *filter.PaymentStatus)
	}
	
	if filter.TableID != nil {
		argCount++
		query += fmt.Sprintf(" AND o.table_id = $%d", argCount)
		args = append(args, *filter.TableID)
	}
	
	if filter.DateFrom != nil {
		argCount++
		query += fmt.Sprintf(" AND o.created_at >= $%d", argCount)
		args = append(args, *filter.DateFrom)
	}
	
	if filter.DateTo != nil {
		argCount++
		query += fmt.Sprintf(" AND o.created_at <= $%d", argCount)
		args = append(args, *filter.DateTo)
	}
	
	query += " ORDER BY o.created_at DESC"
	
	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}
	
	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var table models.Table
		
		err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.SessionID,
			&order.TableID,
			&order.Status,
			&order.TotalAmount,
			&order.TaxAmount,
			&order.ServiceCharge,
			&order.GrandTotal,
			&order.PaymentStatus,
			&order.PaymentMethod,
			&order.Notes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&table.ID,
			&table.TableNumber,
			&table.QRCode,
			&table.Status,
			&table.Capacity,
			&table.CreatedAt,
			&table.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		order.Table = &table
		orders = append(orders, order)
	}
	
	return orders, rows.Err()
}

func (r *orderRepository) UpdateStatus(id int, status models.OrderStatus) error {
	query := `UPDATE orders SET status = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, status)
	return err
}

func (r *orderRepository) UpdatePaymentStatus(id int, status models.PaymentStatus, method *string) error {
	query := `UPDATE orders SET payment_status = $2, payment_method = $3 WHERE id = $1`
	_, err := r.db.Exec(query, id, status, method)
	return err
}

func (r *orderRepository) GetOrderItems(orderID int) ([]models.OrderItem, error) {
	query := `
		SELECT 
			oi.id, oi.order_id, oi.menu_item_id, oi.quantity,
			oi.unit_price, oi.subtotal, oi.notes, oi.status,
			oi.created_at, oi.updated_at,
			mi.id, mi.category_id, mi.name, mi.description, mi.price,
			mi.image_url, mi.image_360_url, mi.video_url, mi.is_available,
			mi.preparation_time, mi.stock_quantity, mi.created_at, mi.updated_at
		FROM order_items oi
		LEFT JOIN menu_items mi ON oi.menu_item_id = mi.id
		WHERE oi.order_id = $1
		ORDER BY oi.id
	`
	
	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		var menuItem models.MenuItem
		
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.MenuItemID,
			&item.Quantity,
			&item.UnitPrice,
			&item.Subtotal,
			&item.Notes,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
			&menuItem.ID,
			&menuItem.CategoryID,
			&menuItem.Name,
			&menuItem.Description,
			&menuItem.Price,
			&menuItem.ImageURL,
			&menuItem.Image360URL,
			&menuItem.VideoURL,
			&menuItem.IsAvailable,
			&menuItem.PreparationTime,
			&menuItem.StockQuantity,
			&menuItem.CreatedAt,
			&menuItem.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		item.MenuItem = &menuItem
		items = append(items, item)
	}
	
	return items, rows.Err()
}

func (r *orderRepository) UpdateOrderItemStatus(id int, status models.OrderItemStatus) error {
	query := `UPDATE order_items SET status = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, status)
	return err
}

func (r *orderRepository) generateOrderNumber() string {
	// Format: ORD-YYYYMMDD-XXXXX
	now := time.Now()
	dateStr := now.Format("20060102")
	
	// Get today's order count
	var count int
	query := `SELECT COUNT(*) FROM orders WHERE DATE(created_at) = DATE($1)`
	r.db.QueryRow(query, now).Scan(&count)
	
	return fmt.Sprintf("ORD-%s-%05d", dateStr, count+1)
}
