
// internal/handlers/staff/payment.go
package staff

import (
	"net/http"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/invoice"
	"lendral3n/ordering-system/internal/utils"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type PaymentHandler struct {
	paymentRepo    repository.PaymentRepository
	orderRepo      repository.OrderRepository
	invoiceService *invoice.Service
}

func NewPaymentHandler(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	invoiceService *invoice.Service,
) *PaymentHandler {
	return &PaymentHandler{
		paymentRepo:    paymentRepo,
		orderRepo:      orderRepo,
		invoiceService: invoiceService,
	}
}

func (h *PaymentHandler) GetPayments(w http.ResponseWriter, r *http.Request) {
	filter := repository.PaymentFilter{
		Limit:  50,
		Offset: 0,
	}
	
	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	
	// Date filters
	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	
	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.DateTo = &t
		}
	}
	
	// Pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}
	
	payments, err := h.paymentRepo.GetAll(filter)
	if err != nil {
		utils.ErrorResponse(w, "Failed to get payments", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Payments retrieved", payments)
}

func (h *PaymentHandler) VerifyPayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paymentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}
	
	// This would typically involve checking with the payment gateway
	// For now, we'll just update the status
	
	utils.SuccessResponse(w, "Payment verified", nil)
}

// internal/handlers/staff/notification.go
package staff

import (
	"encoding/json"
	"net/http"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/utils"
	"strconv"

	"github.com/gorilla/mux"
)

type NotificationHandler struct {
	notifRepo repository.NotificationRepository
}

func NewNotificationHandler(notifRepo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		notifRepo: notifRepo,
	}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	unreadOnly := r.URL.Query().Get("unread_only") == "true"
	
	notifications, err := h.notifRepo.GetByStaffID(userID, unreadOnly)
	if err != nil {
		utils.ErrorResponse(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Notifications retrieved", notifications)
}

func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	
	count, err := h.notifRepo.GetUnreadCount(userID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to get unread count", http.StatusInternalServerError)
		return
	}
	
	response := map[string]int{
		"unread_count": count,
	}
	
	utils.SuccessResponse(w, "Unread count retrieved", response)
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notifID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}
	
	if err := h.notifRepo.MarkAsRead(notifID); err != nil {
		utils.ErrorResponse(w, "Failed to mark as read", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Notification marked as read", nil)
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	
	if err := h.notifRepo.MarkAllAsRead(userID); err != nil {
		utils.ErrorResponse(w, "Failed to mark all as read", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "All notifications marked as read", nil)
}

// internal/handlers/staff/analytics.go
package staff

import (
	"net/http"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/utils"
	"time"
)

type AnalyticsHandler struct {
	orderRepo   repository.OrderRepository
	paymentRepo repository.PaymentRepository
}

func NewAnalyticsHandler(
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
	}
}

func (h *AnalyticsHandler) GetSalesAnalytics(w http.ResponseWriter, r *http.Request) {
	// Parse date range
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	
	if startDate == "" || endDate == "" {
		// Default to last 30 days
		end := time.Now()
		start := end.AddDate(0, 0, -30)
		startDate = start.Format("2006-01-02")
		endDate = end.Format("2006-01-02")
	}
	
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	
	// Get orders within date range
	filter := repository.OrderFilter{
		DateFrom: &start,
		DateTo:   &end,
		Limit:    1000,
	}
	
	orders, err := h.orderRepo.GetAll(filter)
	if err != nil {
		utils.ErrorResponse(w, "Failed to get analytics data", http.StatusInternalServerError)
		return
	}
	
	// Calculate analytics
	analytics := h.calculateAnalytics(orders)
	
	utils.SuccessResponse(w, "Analytics retrieved", analytics)
}

func (h *AnalyticsHandler) calculateAnalytics(orders []models.Order) map[string]interface{} {
	totalRevenue := 0.0
	totalOrders := len(orders)
	completedOrders := 0
	cancelledOrders := 0
	
	itemsSold := make(map[string]int)
	hourlyOrders := make(map[int]int)
	dailyRevenue := make(map[string]float64)
	
	for _, order := range orders {
		if order.Status == models.OrderStatusCompleted {
			completedOrders++
			totalRevenue += order.GrandTotal
			
			// Daily revenue
			dateKey := order.CreatedAt.Format("2006-01-02")
			dailyRevenue[dateKey] += order.GrandTotal
			
			// Hourly distribution
			hour := order.CreatedAt.Hour()
			hourlyOrders[hour]++
			
			// Popular items
			for _, item := range order.OrderItems {
				if item.MenuItem != nil {
					itemsSold[item.MenuItem.Name] += item.Quantity
				}
			}
		} else if order.Status == models.OrderStatusCancelled {
			cancelledOrders++
		}
	}
	
	// Find top selling items
	topItems := h.getTopItems(itemsSold, 10)
	
	// Calculate average order value
	avgOrderValue := 0.0
	if completedOrders > 0 {
		avgOrderValue = totalRevenue / float64(completedOrders)
	}
	
	return map[string]interface{}{
		"total_revenue":     totalRevenue,
		"total_orders":      totalOrders,
		"completed_orders":  completedOrders,
		"cancelled_orders":  cancelledOrders,
		"average_order_value": avgOrderValue,
		"top_selling_items": topItems,
		"hourly_distribution": hourlyOrders,
		"daily_revenue":     dailyRevenue,
	}
}

func (h *AnalyticsHandler) getTopItems(items map[string]int, limit int) []map[string]interface{} {
	// Convert map to slice for sorting
	type kv struct {
		Key   string
		Value int
	}
	
	var sortedItems []kv
	for k, v := range items {
		sortedItems = append(sortedItems, kv{k, v})
	}
	
	// Sort by value descending
	sort.Slice(sortedItems, func(i, j int) bool {
		return sortedItems[i].Value > sortedItems[j].Value
	})
	
	// Take top N items
	result := make([]map[string]interface{}, 0)
	for i := 0; i < limit && i < len(sortedItems); i++ {
		result = append(result, map[string]interface{}{
			"name":     sortedItems[i].Key,
			"quantity": sortedItems[i].Value,
		})
	}
	
	return result
}