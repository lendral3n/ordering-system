// internal/handlers/staff/handlers.go
package staff

import (
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/auth"
	"lendral3n/ordering-system/internal/services/invoice"
	"lendral3n/ordering-system/internal/services/media"
	"lendral3n/ordering-system/internal/services/notification"
)

type Handlers struct {
	Auth         *AuthHandler
	Order        *OrderHandler
	Menu         *MenuHandler
	Payment      *PaymentHandler
	Notification *NotificationHandler
	Analytics    *AnalyticsHandler
}

func NewHandlers(
	staffRepo repository.StaffRepository,
	menuRepo repository.MenuRepository,
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	mediaRepo repository.MediaRepository,
	authService *auth.Service,
	mediaService *media.Service,
	invoiceService *invoice.Service,
	notifService *notification.Service,
) *Handlers {
	return &Handlers{
		Auth:         NewAuthHandler(staffRepo, authService),
		Order:        NewOrderHandler(orderRepo, invoiceService, notifService),
		Menu:         NewMenuHandler(menuRepo, mediaRepo, mediaService),
		Payment:      NewPaymentHandler(paymentRepo, orderRepo, invoiceService),
		Notification: NewNotificationHandler(repository.NewNotificationRepository(staffRepo.(*repository.staffRepository).db)),
		Analytics:    NewAnalyticsHandler(orderRepo, paymentRepo),
	}
}

// internal/handlers/staff/auth.go
package staff

import (
	"encoding/json"
	"net/http"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/auth"
	"lendral3n/ordering-system/internal/utils"
)

type AuthHandler struct {
	staffRepo   repository.StaffRepository
	authService *auth.Service
}

func NewAuthHandler(staffRepo repository.StaffRepository, authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		staffRepo:   staffRepo,
		authService: authService,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string                 `json:"token"`
	User     map[string]interface{} `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate request
	if req.Username == "" || req.Password == "" {
		utils.ErrorResponse(w, "Username and password required", http.StatusBadRequest)
		return
	}
	
	// Get staff by username
	staff, err := h.staffRepo.GetByUsername(req.Username)
	if err != nil {
		utils.ErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Check if staff is active
	if !staff.IsActive {
		utils.ErrorResponse(w, "Account is disabled", http.StatusUnauthorized)
		return
	}
	
	// Verify password
	if err := h.authService.ComparePassword(staff.PasswordHash, req.Password); err != nil {
		utils.ErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Generate JWT token
	token, err := h.authService.GenerateToken(staff)
	if err != nil {
		utils.ErrorResponse(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	
	// Update last login
	h.staffRepo.UpdateLastLogin(staff.ID)
	
	response := LoginResponse{
		Token: token,
		User: map[string]interface{}{
			"id":        staff.ID,
			"username":  staff.Username,
			"email":     staff.Email,
			"full_name": staff.FullName,
			"role":      staff.Role,
		},
	}
	
	utils.SuccessResponse(w, "Login successful", response)
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("user_id").(int)
	
	staff, err := h.staffRepo.GetByID(userID)
	if err != nil {
		utils.ErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}
	
	user := map[string]interface{}{
		"id":         staff.ID,
		"username":   staff.Username,
		"email":      staff.Email,
		"full_name":  staff.FullName,
		"role":       staff.Role,
		"is_active":  staff.IsActive,
		"last_login": staff.LastLogin,
		"created_at": staff.CreatedAt,
	}
	
	utils.SuccessResponse(w, "Profile retrieved", user)
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate new password
	if len(req.NewPassword) < 6 {
		utils.ErrorResponse(w, "New password must be at least 6 characters", http.StatusBadRequest)
		return
	}
	
	// Get staff
	staff, err := h.staffRepo.GetByID(userID)
	if err != nil {
		utils.ErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Verify current password
	if err := h.authService.ComparePassword(staff.PasswordHash, req.CurrentPassword); err != nil {
		utils.ErrorResponse(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}
	
	// Hash new password
	hashedPassword, err := h.authService.HashPassword(req.NewPassword)
	if err != nil {
		utils.ErrorResponse(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	
	// Update password
	staff.PasswordHash = hashedPassword
	if err := h.staffRepo.Update(staff); err != nil {
		utils.ErrorResponse(w, "Failed to update password", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Password changed successfully", nil)
}

// internal/handlers/staff/order.go
package staff

import (
	"encoding/json"
	"net/http"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/invoice"
	"lendral3n/ordering-system/internal/services/notification"
	"lendral3n/ordering-system/internal/utils"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	orderRepo      repository.OrderRepository
	invoiceService *invoice.Service
	notifService   *notification.Service
}

func NewOrderHandler(
	orderRepo repository.OrderRepository,
	invoiceService *invoice.Service,
	notifService *notification.Service,
) *OrderHandler {
	return &OrderHandler{
		orderRepo:      orderRepo,
		invoiceService: invoiceService,
		notifService:   notifService,
	}
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := repository.OrderFilter{
		Limit:  50,
		Offset: 0,
	}
	
	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		orderStatus := models.OrderStatus(status)
		filter.Status = &orderStatus
	}
	
	// Payment status filter
	if paymentStatus := r.URL.Query().Get("payment_status"); paymentStatus != "" {
		ps := models.PaymentStatus(paymentStatus)
		filter.PaymentStatus = &ps
	}
	
	// Table ID filter
	if tableIDStr := r.URL.Query().Get("table_id"); tableIDStr != "" {
		if tableID, err := strconv.Atoi(tableIDStr); err == nil {
			filter.TableID = &tableID
		}
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
	
	// Get orders
	orders, err := h.orderRepo.GetAll(filter)
	if err != nil {
		utils.ErrorResponse(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Orders retrieved", orders)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	
	order, err := h.orderRepo.GetByID(orderID)
	if err != nil {
		utils.ErrorResponse(w, "Order not found", http.StatusNotFound)
		return
	}
	
	utils.SuccessResponse(w, "Order retrieved", order)
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate status
	status := models.OrderStatus(req.Status)
	validStatuses := []models.OrderStatus{
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
		if status == s {
			isValid = true
			break
		}
	}
	
	if !isValid {
		utils.ErrorResponse(w, "Invalid order status", http.StatusBadRequest)
		return
	}
	
	// Get order
	order, err := h.orderRepo.GetByID(orderID)
	if err != nil {
		utils.ErrorResponse(w, "Order not found", http.StatusNotFound)
		return
	}
	
	// Update status
	if err := h.orderRepo.UpdateStatus(orderID, status); err != nil {
		utils.ErrorResponse(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}
	
	// Update order object for notification
	order.Status = status
	
	// Send notification
	go h.notifService.NotifyOrderStatusUpdate(order)
	
	utils.SuccessResponse(w, "Order status updated", nil)
}

func (h *OrderHandler) UpdateOrderItemStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.Atoi(vars["item_id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid item ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate status
	status := models.OrderItemStatus(req.Status)
	validStatuses := []models.OrderItemStatus{
		models.OrderItemStatusPending,
		models.OrderItemStatusPreparing,
		models.OrderItemStatusReady,
		models.OrderItemStatusServed,
		models.OrderItemStatusCancelled,
	}
	
	isValid := false
	for _, s := range validStatuses {
		if status == s {
			isValid = true
			break
		}
	}
	
	if !isValid {
		utils.ErrorResponse(w, "Invalid item status", http.StatusBadRequest)
		return
	}
	
	// Update status
	if err := h.orderRepo.UpdateOrderItemStatus(itemID, status); err != nil {
		utils.ErrorResponse(w, "Failed to update item status", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Item status updated", nil)
}

func (h *OrderHandler) GenerateInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	
	// Generate invoice PDF
	pdfData, invoiceNumber, err := h.invoiceService.GenerateInvoice(orderID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to generate invoice", http.StatusInternalServerError)
		return
	}
	
	// Set response headers
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=invoice_%s.pdf", invoiceNumber))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfData)))
	
	// Write PDF data
	w.Write(pdfData)
}