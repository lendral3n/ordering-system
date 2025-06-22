package handlers

import (
	"fmt"
	"lendral3n/ordering-system/internal/models"

	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) HandleMidtransNotification(c *fiber.Ctx) error {
	// Parse notification
	body := c.Body()
	notif, err := h.MidtransService.ParseNotification(body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid notification",
		})
	}

	// Get payment record
	var payment models.Payment
	if err := h.DB.Where("midtrans_order_id = ?", notif.OrderID).First(&payment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Payment not found",
		})
	}

	// Get order
	var order models.Order
	if err := h.DB.Where("order_number = ?", notif.OrderID).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Order not found",
		})
	}

	// Update payment record
	payment.MidtransTransactionID = &notif.TransactionID
	payment.TransactionStatus = &notif.TransactionStatus
	payment.PaymentType = &notif.PaymentType
	payment.StatusMessage = &notif.StatusMessage

	if transTime, err := notif.GetTransactionTime(); err == nil {
		payment.TransactionTime = transTime
	}

	vaNumber := notif.GetVANumber()
	bank := notif.GetBank()
	payment.VANumber = &vaNumber
	payment.Bank = &bank
	payment.FraudStatus = &notif.FraudStatus

	if err := h.DB.Save(&payment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update payment",
		})
	}

	// Update order payment status based on transaction status
	switch notif.TransactionStatus {
	case "capture", "settlement":
		// Payment successful
		h.DB.Model(&order).Updates(map[string]interface{}{
			"payment_status": models.PaymentStatusPaid,
			"payment_method": notif.PaymentType,
		})

		// Create notification
		notification := models.Notification{
			OrderID: &order.ID,
			Type:    models.NotificationPaymentReceived,
			Message: fmt.Sprintf("Payment received for order #%s", order.OrderNumber),
		}
		h.DB.Create(&notification)

		go h.NotificationHub.BroadcastPaymentReceived(&payment, &order)

	case "pending":
		// Payment pending
		h.DB.Model(&order).Update("payment_status", models.PaymentStatusPending)

	case "deny", "cancel", "expire":
		// Payment failed
		h.DB.Model(&order).Update("payment_status", models.PaymentStatusFailed)

	case "refund":
		// Payment refunded
		h.DB.Model(&order).Update("payment_status", models.PaymentStatusRefunded)
	}

	// Send success response
	return c.SendString("OK")
}
