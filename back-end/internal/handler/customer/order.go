
// internal/handlers/customer/order.go
package customer

import (
	"encoding/json"
	"net/http"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/notification"
	"lendral3n/ordering-system/internal/utils"
	"strconv"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	orderRepo    repository.OrderRepository
	menuRepo     repository.MenuRepository
	notifService *notification.Service
}

func NewOrderHandler(
	orderRepo repository.OrderRepository,
	menuRepo repository.MenuRepository,
	notifService *notification.Service,
) *OrderHandler {
	return &OrderHandler{
		orderRepo:    orderRepo,
		menuRepo:     menuRepo,
		notifService: notifService,
	}
}

type CreateOrderRequest struct {
	SessionToken string                   `json:"session_token"`
	Items        []CreateOrderItemRequest `json:"items"`
	Notes        string                   `json:"notes"`
}

type CreateOrderItemRequest struct {
	MenuItemID int    `json:"menu_item_id"`
	Quantity   int    `json:"quantity"`
	Notes      string `json:"notes"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Get session from header
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		utils.ErrorResponse(w, "Session token required", http.StatusUnauthorized)
		return
	}
	
	// Validate session
	sessionRepo := repository.NewSessionRepository(h.orderRepo.(*repository.orderRepository).db)
	session, err := sessionRepo.GetByToken(sessionToken)
	if err != nil || session.EndedAt != nil {
		utils.ErrorResponse(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}
	
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate request
	if len(req.Items) == 0 {
		utils.ErrorResponse(w, "Order must contain at least one item", http.StatusBadRequest)
		return
	}
	
	// Get config for tax and service percentages
	taxPercentage := 0.10    // 10%
	servicePercentage := 0.05 // 5%
	
	// Validate menu items and calculate totals
	var totalAmount float64
	orderItems := make([]models.OrderItem, 0, len(req.Items))
	
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			utils.ErrorResponse(w, "Invalid quantity", http.StatusBadRequest)
			return
		}
		
		menuItem, err := h.menuRepo.GetMenuItemByID(item.MenuItemID)
		if err != nil {
			utils.ErrorResponse(w, "Menu item not found", http.StatusBadRequest)
			return
		}
		
		if !menuItem.IsAvailable {
			utils.ErrorResponse(w, fmt.Sprintf("Menu item '%s' is not available", menuItem.Name), http.StatusBadRequest)
			return
		}
		
		// Check stock if tracked
		if menuItem.StockQuantity.Valid && menuItem.StockQuantity.Int64 < int64(item.Quantity) {
			utils.ErrorResponse(w, fmt.Sprintf("Insufficient stock for '%s'", menuItem.Name), http.StatusBadRequest)
			return
		}
		
		subtotal := menuItem.Price * float64(item.Quantity)
		totalAmount += subtotal
		
		orderItems = append(orderItems, models.OrderItem{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			UnitPrice:  menuItem.Price,
			Subtotal:   subtotal,
			Notes:      &item.Notes,
			Status:     models.OrderItemStatusPending,
			MenuItem:   menuItem,
		})
	}
	
	// Calculate tax and service charge
	taxAmount := totalAmount * taxPercentage
	serviceCharge := totalAmount * servicePercentage
	grandTotal := totalAmount + taxAmount + serviceCharge
	
	// Create order
	order := &models.Order{
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
	
	createdOrder, err := h.orderRepo.Create(order)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create order", http.StatusInternalServerError)
		return
	}
	
	// Send notification to staff
	go h.notifService.NotifyNewOrder(createdOrder)
	
	utils.SuccessResponse(w, "Order created successfully", createdOrder)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Validate session
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		utils.ErrorResponse(w, "Session token required", http.StatusUnauthorized)
		return
	}
	
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
	
	// Verify order belongs to session
	sessionRepo := repository.NewSessionRepository(h.orderRepo.(*repository.orderRepository).db)
	session, err := sessionRepo.GetByToken(sessionToken)
	if err != nil || session.ID != order.SessionID {
		utils.ErrorResponse(w, "Unauthorized", http.StatusForbidden)
		return
	}
	
	utils.SuccessResponse(w, "Order retrieved", order)
}

func (h *OrderHandler) GetOrdersBySession(w http.ResponseWriter, r *http.Request) {
	// Validate session
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		utils.ErrorResponse(w, "Session token required", http.StatusUnauthorized)
		return
	}
	
	sessionRepo := repository.NewSessionRepository(h.orderRepo.(*repository.orderRepository).db)
	session, err := sessionRepo.GetByToken(sessionToken)
	if err != nil {
		utils.ErrorResponse(w, "Invalid session", http.StatusUnauthorized)
		return
	}
	
	orders, err := h.orderRepo.GetBySessionID(session.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}
	
	utils.SuccessResponse(w, "Orders retrieved", orders)
}

func (h *OrderHandler) RequestAssistance(w http.ResponseWriter, r *http.Request) {
	// Validate session
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		utils.ErrorResponse(w, "Session token required", http.StatusUnauthorized)
		return
	}
	
	sessionRepo := repository.NewSessionRepository(h.orderRepo.(*repository.orderRepository).db)
	session, err := sessionRepo.GetByToken(sessionToken)
	if err != nil || session.EndedAt != nil {
		utils.ErrorResponse(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}
	
	// Send notification to staff
	go h.notifService.NotifyAssistanceRequest(session.TableID, session.Table.TableNumber)
	
	utils.SuccessResponse(w, "Assistance request sent", nil)
}
