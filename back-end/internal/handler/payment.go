package handlers

import (
	"fmt"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/services/payment"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CreatePaymentRequest struct {
	OrderID       int    `json:"order_id" validate:"required"`
	CustomerEmail string `json:"customer_email,omitempty"`
}

type CreatePaymentResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

func (h *Handlers) CreatePayment(c *fiber.Ctx) error {
	// Validate session
	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Session token required",
		})
	}

	var req CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Get order
	var order models.Order
	if err := h.DB.Preload("Table").Preload("OrderItems.MenuItem").First(&order, req.OrderID).Error; err != nil {
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

	// Check if order is already paid
	if order.PaymentStatus == models.PaymentStatusPaid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Order already paid",
		})
	}

	// Check if payment already exists
	var existingPayment models.Payment
	err := h.DB.Where("order_id = ?", order.ID).First(&existingPayment).Error
	if err == nil && existingPayment.TransactionStatus != nil {
		status := *existingPayment.TransactionStatus
		if status == "pending" || status == "settlement" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Payment already in progress",
			})
		}
	}

	// Get customer info from session
	customerName := "Guest"
	customerPhone := ""
	if session.CustomerName != nil {
		customerName = *session.CustomerName
	}
	if session.CustomerPhone != nil {
		customerPhone = *session.CustomerPhone
	}

	// Create Midtrans transaction
	transReq := payment.CreateTransactionRequest{
		Order:         &order,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
		CustomerEmail: req.CustomerEmail,
	}

	transResp, err := h.MidtransService.CreateTransaction(transReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create payment",
		})
	}

	// Create payment record
	paymentRecord := models.Payment{
		OrderID:           order.ID,
		MidtransOrderID:   order.OrderNumber,
		TransactionStatus: &[]string{"pending"}[0],
		GrossAmount:       order.GrandTotal,
		Currency:          "IDR",
	}

	if err := h.DB.Create(&paymentRecord).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to save payment record",
		})
	}

	// Update order payment status
	h.DB.Model(&order).Update("payment_status", models.PaymentStatusPending)

	response := CreatePaymentResponse{
		Token:       transResp.Token,
		RedirectURL: transResp.RedirectURL,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Payment created",
		"data":    response,
	})
}

func (h *Handlers) GetPaymentStatus(c *fiber.Ctx) error {
	// Validate session
	sessionToken := c.Get("X-Session-Token")
	if sessionToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Session token required",
		})
	}

	orderID, err := strconv.Atoi(c.Params("order_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid order ID",
		})
	}

	// Get order
	var order models.Order
	if err := h.DB.First(&order, orderID).Error; err != nil {
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

	// Get payment
	var payment models.Payment
	if err := h.DB.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Payment not found",
		})
	}

	// Check transaction status from Midtrans
	status, err := h.MidtransService.GetTransactionStatus(order.OrderNumber)
	if err == nil && status != nil {
		// Update payment record if status changed
		if payment.TransactionStatus == nil || *payment.TransactionStatus != status.TransactionStatus {
			payment.TransactionStatus = &status.TransactionStatus
			payment.MidtransTransactionID = &status.TransactionID
			payment.PaymentType = &status.PaymentType

			if status.VANumbers != nil && len(status.VANumbers) > 0 {
				payment.VANumber = &status.VANumbers[0].VANumber
				payment.Bank = &status.VANumbers[0].Bank
			}

			h.DB.Save(&payment)

			// Update order payment status
			if status.TransactionStatus == "settlement" || status.TransactionStatus == "capture" {
				h.DB.Model(&order).Updates(map[string]interface{}{
					"payment_status": models.PaymentStatusPaid,
					"payment_method": status.PaymentType,
				})

				// Create notification
				notification := models.Notification{
					OrderID: &order.ID,
					Type:    models.NotificationPaymentReceived,
					Message: fmt.Sprintf("Payment received for order #%s", order.OrderNumber),
				}
				h.DB.Create(&notification)

				// Send notifications
				go h.NotificationHub.BroadcastPaymentReceived(&payment, &order)
			} else if status.TransactionStatus == "expire" || status.TransactionStatus == "cancel" {
				h.DB.Model(&order).Update("payment_status", models.PaymentStatusFailed)
			}
		}
	}

	response := map[string]interface{}{
		"order_id":           order.ID,
		"payment_status":     order.PaymentStatus,
		"transaction_status": payment.TransactionStatus,
		"payment_type":       payment.PaymentType,
		"va_number":          payment.VANumber,
		"bank":               payment.Bank,
		"gross_amount":       payment.GrossAmount,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Payment status retrieved",
		"data":    response,
	})
}

func (h *Handlers) HandlePaymentFinish(c *fiber.Ctx) error {
	// This endpoint is called when user finishes payment on Midtrans page
	orderID := c.Query("order_id")
	statusCode := c.Query("status_code")
	transactionStatus := c.Query("transaction_status")

	// Build redirect URL based on status
	redirectURL := "/payment-success"
	if transactionStatus == "deny" || transactionStatus == "expire" || transactionStatus == "cancel" {
		redirectURL = "/payment-failed"
	}

	redirectURL += "?order_id=" + orderID + "&status=" + transactionStatus + "&code=" + statusCode

	return c.Redirect(redirectURL)
}

// Staff endpoints
func (h *Handlers) GetPayments(c *fiber.Ctx) error {
	// Parse query parameters
	status := c.Query("status")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	query := h.DB.Preload("Order.Table")

	// Apply filters
	if status != "" {
		query = query.Where("transaction_status = ?", status)
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

	var payments []models.Payment
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get payments",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Payments retrieved",
		"data":    payments,
	})
}

func (h *Handlers) VerifyPayment(c *fiber.Ctx) error {
	paymentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid payment ID",
		})
	}

	// Get payment
	var payment models.Payment
	if err := h.DB.Preload("Order").First(&payment, paymentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Payment not found",
		})
	}

	// Check transaction status from Midtrans
	status, err := h.MidtransService.GetTransactionStatus(payment.MidtransOrderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to verify payment",
		})
	}

	// Update payment record
	payment.TransactionStatus = &status.TransactionStatus
	payment.MidtransTransactionID = &status.TransactionID
	payment.PaymentType = &status.PaymentType
	h.DB.Save(&payment)

	// Update order if payment successful
	if status.TransactionStatus == "settlement" || status.TransactionStatus == "capture" {
		h.DB.Model(&models.Order{}).Where("id = ?", payment.OrderID).Updates(map[string]interface{}{
			"payment_status": models.PaymentStatusPaid,
			"payment_method": status.PaymentType,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Payment verified",
		"data":    payment,
	})
}
