package routes

import (
	"lendral3n/ordering-system/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h *handlers.Handlers) {
	// API routes
	api := app.Group("/api")

	// Customer routes
	customer := api.Group("/customer")
	
	// Session routes
	customer.Post("/session/start", h.StartSession)
	customer.Get("/session", h.GetSession)
	customer.Post("/session/end", h.EndSession)
	
	// Menu routes (public)
	customer.Get("/menu/categories", h.GetCategories)
	customer.Get("/menu/items", h.GetMenuItems)
	customer.Get("/menu/items/:id", h.GetMenuItem)
	customer.Get("/menu/items/:id/360", h.GetMenu360View)
	
	// Order routes (require session)
	customer.Post("/orders", h.CreateOrder)
	customer.Get("/orders/:id", h.GetOrder)
	customer.Get("/orders/session", h.GetOrdersBySession)
	customer.Post("/assistance", h.RequestAssistance)
	
	// Payment routes
	customer.Post("/payment/create", h.CreatePayment)
	customer.Get("/payment/:order_id/status", h.GetPaymentStatus)
	customer.Get("/payment/finish", h.HandlePaymentFinish)
	customer.Post("/payment/finish", h.HandlePaymentFinish)

	// Staff routes
	staff := api.Group("/staff")
	
	// Order management
	staff.Get("/orders", h.GetOrders)
	staff.Get("/orders/:id", h.GetOrder)
	staff.Put("/orders/:id/status", h.UpdateOrderStatus)
	staff.Put("/orders/items/:item_id/status", h.UpdateOrderItemStatus)
	
	// Menu management
	staff.Get("/menu/categories", h.GetCategories)
	staff.Post("/menu/categories", h.CreateCategory)
	staff.Put("/menu/categories/:id", h.UpdateCategory)
	
	staff.Get("/menu/items", h.GetMenuItems)
	staff.Post("/menu/items", h.CreateMenuItem)
	staff.Put("/menu/items/:id", h.UpdateMenuItem)
	staff.Delete("/menu/items/:id", h.DeleteMenuItem)
	staff.Post("/menu/items/:id/media", h.UploadMedia)
	staff.Put("/menu/items/:id/stock", h.UpdateStock)
	
	// Payment routes
	staff.Get("/payments", h.GetPayments)
	staff.Put("/payments/:id/verify", h.VerifyPayment)
	
	// Notification routes
	staff.Get("/notifications", h.GetNotifications)
	staff.Get("/notifications/unread-count", h.GetUnreadCount)
	staff.Put("/notifications/:id/read", h.MarkAsRead)
	staff.Put("/notifications/read-all", h.MarkAllAsRead)
	
	// Analytics routes
	staff.Get("/analytics/sales", h.GetSalesAnalytics)
	staff.Get("/analytics/tables", h.GetTableAnalytics)
	staff.Get("/analytics/menu", h.GetMenuPerformance)

	// Webhook routes
	webhook := api.Group("/webhook")
	webhook.Post("/midtrans", h.HandleMidtransNotification)
}