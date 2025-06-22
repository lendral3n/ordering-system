
// internal/handlers/customer/payment.go
package customer

import (
	"encoding/json"
	"net/http"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/notification"
	"lendral3n/ordering-system/internal/services/payment"
	"lendral3n/ordering-system/internal/utils"
	"strconv"

	"github.com/gorilla/mux"
)

type PaymentHandler struct {
	orderRepo      repository.OrderRepository
	paymentRepo    repository.PaymentRepository
	paymentService *payment.MidtransService
	notifService   *notification.Service
}

func NewPaymentHandler(
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	paymentService *payment.MidtransService,
	notifService *notification.Service,
) *PaymentHandler {
	return &PaymentHandler{
		orderRepo:      orderRepo,
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
		notifService:   notifService,
	}
}

type CreatePaymentRequest struct {
	OrderID       int    `json:"order_id"`
	CustomerEmail string `json:"customer_email,omitempty"`
}

type CreatePaymentResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	// Validate session
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		utils.ErrorResponse(w, "Session token required", http.StatusUnauthorized)
		return
	}
	
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Get order
	order, err := h.orderRepo.GetByID(req.OrderID)
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
	
	// Check if order is already paid
	if order.PaymentStatus == models.PaymentStatusPaid {
		utils.ErrorResponse(w, "Order already paid", http.StatusBadRequest)
		return
	}
	
	// Check if payment already exists
	existingPayment, _ := h.paymentRepo.GetByOrderID(order.ID)
	if existingPayment != nil && existingPayment.TransactionStatus != nil {
		status := *existingPayment.TransactionStatus
		if status == "pending" || status == "settlement" {
			utils.ErrorResponse(w, "Payment already in progress", http.StatusBadRequest)
			return
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
		Order:         order,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
		CustomerEmail: req.CustomerEmail,
	}
	
	transResp, err := h.paymentService.CreateTransaction(transReq)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create payment", http.StatusInternalServerError)
		return
	}
	
	// Create payment record
	paymentRecord := &models.Payment{
		OrderID:           order.ID,
		MidtransOrderID:   order.OrderNumber,
		TransactionStatus: utils.StringPtr("pending"),
		GrossAmount:       order.GrandTotal,
		Currency:          "IDR",
	}
	
	if err := h.paymentRepo.Create(paymentRecord); err != nil {
		utils.ErrorResponse(w, "Failed to save payment record", http.StatusInternalServerError)
		return
	}
	
	// Update order payment status
	h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusPending, nil)
	
	response := CreatePaymentResponse{
		Token:       transResp.Token,
		RedirectURL: transResp.RedirectURL,
	}
	
	utils.SuccessResponse(w, "Payment created", response)
}

func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	// Validate session
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		utils.ErrorResponse(w, "Session token required", http.StatusUnauthorized)
		return
	}
	
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["order_id"])
	if err != nil {
		utils.ErrorResponse(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	
	// Get order
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
	
	// Get payment
	payment, err := h.paymentRepo.GetByOrderID(orderID)
	if err != nil {
		utils.ErrorResponse(w, "Payment not found", http.StatusNotFound)
		return
	}
	
	// Check transaction status from Midtrans
	status, err := h.paymentService.GetTransactionStatus(order.OrderNumber)
	if err == nil && status != nil {
		// Update payment record if status changed
		if payment.TransactionStatus == nil || *payment.TransactionStatus != status.TransactionStatus {
			payment.TransactionStatus = &status.TransactionStatus
			payment.TransactionID = &status.TransactionID
			payment.PaymentType = &status.PaymentType
			
			if status.VANumbers != nil && len(status.VANumbers) > 0 {
				payment.VANumber = &status.VANumbers[0].VANumber
				payment.Bank = &status.VANumbers[0].Bank
			}
			
			h.paymentRepo.Update(payment)
			
			// Update order payment status
			if status.TransactionStatus == "settlement" || status.TransactionStatus == "capture" {
				h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusPaid, &status.PaymentType)
				
				// Send notifications
				go h.notifService.NotifyPaymentReceived(payment, order)
			} else if status.TransactionStatus == "expire" || status.TransactionStatus == "cancel" {
				h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusFailed, nil)
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
	
	utils.SuccessResponse(w, "Payment status retrieved", response)
}

func (h *PaymentHandler) HandlePaymentFinish(w http.ResponseWriter, r *http.Request) {
	// This endpoint is called when user finishes payment on Midtrans page
	// Usually you would redirect to a success page in your app
	
	orderID := r.URL.Query().Get("order_id")
	statusCode := r.URL.Query().Get("status_code")
	transactionStatus := r.URL.Query().Get("transaction_status")
	
	// Build redirect URL based on status
	redirectURL := "/payment-success"
	if transactionStatus == "deny" || transactionStatus == "expire" || transactionStatus == "cancel" {
		redirectURL = "/payment-failed"
	}
	
	redirectURL += "?order_id=" + orderID + "&status=" + transactionStatus + "&code=" + statusCode
	
	http.Redirect(w, r, redirectURL, http.StatusFound)
}