package handlers

import (
	"fmt"
	"lendral3n/ordering-system/internal/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreateOrderRequest struct {
	SessionToken string                   `json:"session_token"`
	Items        []CreateOrderItemRequest `json:"items" validate:"required,min=1"`
	Notes        string                   `json:"notes"`
}

type CreateOrderItemRequest struct {
	MenuItemID int    `json:"menu_item_id" validate:"required"`
	Quantity   int    `json:"quantity" validate:"required,min=1"`
	Notes      string `json:"notes"`
}

func (h *Handlers) CreateOrder(c *fiber.Ctx) error {
	// Get session from header
	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Session token required",
		})
	}

	// Validate session
	var session models.CustomerSession
	if err := h.DB.Where("session_token = ? AND ended_at IS NULL", sessionToken).First(&session).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid or expired session",
		})
	}

	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate request
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Order must contain at least one item",
		})
	}

	// Create order in transaction
	var order models.Order
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// Calculate totals
		var totalAmount float64
		orderItems := make([]models.OrderItem, 0, len(req.Items))

		for _, item := range req.Items {
			if item.Quantity <= 0 {
				return fiber.NewError(fiber.StatusBadRequest, "Invalid quantity")
			}

			// Get menu item
			var menuItem models.MenuItem
			if err := tx.First(&menuItem, item.MenuItemID).Error; err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "Menu item not found")
			}

			if !menuItem.IsAvailable {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Menu item '%s' is not available", menuItem.Name))
			}

			// Check stock if tracked
			if menuItem.StockQuantity != nil && *menuItem.StockQuantity < item.Quantity {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Insufficient stock for '%s'", menuItem.Name))
			}

			subtotal := menuItem.Price * float64(item.Quantity)
			totalAmount += subtotal

			orderItems = append(orderItems, models.OrderItem{
				MenuItemID: uint(item.MenuItemID),
				Quantity:   item.Quantity,
				UnitPrice:  menuItem.Price,
				Subtotal:   subtotal,
				Notes:      &item.Notes,
				Status:     models.OrderItemStatusPending,
			})
		}

		// Calculate tax and service charge
		taxAmount := totalAmount * h.Config.TaxPercentage / 100
		serviceCharge := totalAmount * h.Config.ServicePercentage / 100
		grandTotal := totalAmount + taxAmount + serviceCharge

		// Create order
		order = models.Order{
			OrderNumber:   generateOrderNumber(),
			SessionID:     session.ID,
			TableID:       session.TableID,
			Status:        models.OrderStatusPending,
			TotalAmount:   totalAmount,
			TaxAmount:     taxAmount,
			ServiceCharge: serviceCharge,
			GrandTotal:    grandTotal,
			PaymentStatus: models.PaymentStatusUnpaid,
			Notes:         &req.Notes,
			OrderItems:    orderItems,
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// Update inventory for tracked items
		for _, item := range orderItems {
			var menuItem models.MenuItem
			tx.First(&menuItem, item.MenuItemID)
			
			if menuItem.StockQuantity != nil {
				// Update stock
				tx.Model(&menuItem).Update("stock_quantity", gorm.Expr("stock_quantity - ?", item.Quantity))
				
				// Create inventory log
				log := models.InventoryLog{
					MenuItemID:     item.MenuItemID,
					QuantityChange: -item.Quantity,
					Reason:         "order_placed",
					OrderItemID:    &item.ID,
				}
				tx.Create(&log)
			}
		}

		// Update table status
		tx.Model(&models.Table{}).Where("id = ?", session.TableID).Update("status", models.TableStatusOccupied)

		// Create notification
		notification := models.Notification{
			OrderID: &order.ID,
			Type:    models.NotificationNewOrder,
			Message: fmt.Sprintf("New order #%s from table %d", order.OrderNumber, session.TableID),
		}
		tx.Create(&notification)

		return nil
	})

	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).JSON(fiber.Map{
				"success": false,
				"error":   e.Message,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create order",
		})
	}

	// Load relations
	h.DB.Preload("OrderItems.MenuItem").Preload("Table").First(&order, order.ID)

	// Send notification to staff
	go h.NotificationHub.BroadcastNewOrder(&order)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Order created successfully",
		"data":    order,
	})
}

func (h *Handlers) GetOrder(c *fiber.Ctx) error {
	// Validate session
	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Session token required",
		})
	}

	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid order ID",
		})
	}

	var order models.Order
	if err := h.DB.Preload("OrderItems.MenuItem").Preload("Table").Preload("Payment").First(&order, orderID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Order not found",
		})
	}

	// Verify order belongs to session
	var session models.CustomerSession
	if err := h.DB.Where("session_token = ?", sessionToken).First(&session).Error; err != nil || session.ID != order.SessionID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Order retrieved",
		"data":    order,
	})
}

func (h *Handlers) GetOrdersBySession(c *fiber.Ctx) error {
	// Validate session
	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Session token required",
		})
	}

	var session models.CustomerSession
	if err := h.DB.Where("session_token = ?", sessionToken).First(&session).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid session",
		})
	}

	var orders []models.Order
	if err := h.DB.Preload("OrderItems.MenuItem").Preload("Table").
		Where("session_id = ?", session.ID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get orders",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Orders retrieved",
		"data":    orders,
	})
}

func (h *Handlers) RequestAssistance(c *fiber.Ctx) error {
	// Validate session
	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Session token required",
		})
	}

	var session models.CustomerSession
	if err := h.DB.Preload("Table").Where("session_token = ? AND ended_at IS NULL", sessionToken).First(&session).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid or expired session",
		})
	}

	// Create notification
	notification := models.Notification{
		Type:    models.NotificationAssistanceRequest,
		Message: fmt.Sprintf("Table %s needs assistance", session.Table.TableNumber),
	}

	if err := h.DB.Create(&notification).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to send assistance request",
		})
	}

	// Send notification to staff
	go h.NotificationHub.BroadcastAssistanceRequest(session.TableID, session.Table.TableNumber)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Assistance request sent",
	})
}

// Staff endpoints
func (h *Handlers) GetOrders(c *fiber.Ctx) error {
	// Parse query parameters
	status := c.Query("status")
	paymentStatus := c.Query("payment_status")
	tableID := c.Query("table_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	query := h.DB.Preload("Table").Preload("OrderItems.MenuItem")

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if paymentStatus != "" {
		query = query.Where("payment_status = ?", paymentStatus)
	}
	if tableID != "" {
		query = query.Where("table_id = ?", tableID)
	}
	if dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var orders []models.Order
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get orders",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Orders retrieved",
		"data":    orders,
	})
}

func (h *Handlers) UpdateOrderStatus(c *fiber.Ctx) error {
	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid order ID",
		})
	}

	var req struct {
		Status string `json:"status" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate status
	validStatuses := []string{
		models.OrderStatusPending,
		models.OrderStatusConfirmed,
		models.OrderStatusPreparing,
		models.OrderStatusReady,
		models.OrderStatusServed,
		models.OrderStatusCompleted,
		models.OrderStatusCancelled,
	}

	isValid := false
	for _, s := range validStatuses {
		if req.Status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid order status",
		})
	}

	// Update status
	result := h.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("status", req.Status)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update order status",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Order not found",
		})
	}

	// Get order for notification
	var order models.Order
	h.DB.Preload("Table").First(&order, orderID)

	// Send notification
	go h.NotificationHub.BroadcastOrderStatusUpdate(&order)

	// If order is ready, create notification
	if req.Status == models.OrderStatusReady {
		notification := models.Notification{
			OrderID: &[]uint{uint(orderID)}[0],
			Type:    models.NotificationOrderReady,
			Message: fmt.Sprintf("Order #%s is ready to serve", order.OrderNumber),
		}
		h.DB.Create(&notification)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Order status updated",
	})
}

func (h *Handlers) UpdateOrderItemStatus(c *fiber.Ctx) error {
	itemID, err := strconv.Atoi(c.Params("item_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid item ID",
		})
	}

	var req struct {
		Status string `json:"status" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate status
	validStatuses := []string{
		models.OrderItemStatusPending,
		models.OrderItemStatusPreparing,
		models.OrderItemStatusReady,
		models.OrderItemStatusServed,
		models.OrderItemStatusCancelled,
	}

	isValid := false
	for _, s := range validStatuses {
		if req.Status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid item status",
		})
	}

	// Update status
	result := h.DB.Model(&models.OrderItem{}).Where("id = ?", itemID).Update("status", req.Status)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update item status",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Order item not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Item status updated",
	})
}

func generateOrderNumber() string {
	// Format: ORD-YYYYMMDD-XXXXX
	now := time.Now()
	dateStr := now.Format("20060102")
	random := fmt.Sprintf("%05d", now.Unix()%100000)
	return fmt.Sprintf("ORD-%s-%s", dateStr, random)
}