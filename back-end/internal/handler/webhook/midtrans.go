// internal/handlers/webhook/midtrans.go
package webhook

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/notification"
	"lendral3n/ordering-system/internal/services/payment"
	"lendral3n/ordering-system/internal/utils"
)

type Handlers struct {
	Midtrans *MidtransWebhookHandler
}

func NewHandlers(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	paymentService *payment.MidtransService,
	notifService *notification.Service,
) *Handlers {
	return &Handlers{
		Midtrans: NewMidtransWebhookHandler(paymentRepo, orderRepo, paymentService, notifService),
	}
}

type MidtransWebhookHandler struct {
	paymentRepo    repository.PaymentRepository
	orderRepo      repository.OrderRepository
	paymentService *payment.MidtransService
	notifService   *notification.Service
}

func NewMidtransWebhookHandler(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	paymentService *payment.MidtransService,
	notifService *notification.Service,
) *MidtransWebhookHandler {
	return &MidtransWebhookHandler{
		paymentRepo:    paymentRepo,
		orderRepo:      orderRepo,
		paymentService: paymentService,
		notifService:   notifService,
	}
}

func (h *MidtransWebhookHandler) HandleNotification(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	
	// Parse notification
	notif, err := h.paymentService.ParseNotification(body)
	if err != nil {
		utils.ErrorResponse(w, "Invalid notification", http.StatusBadRequest)
		return
	}
	
	// Get payment record
	payment, err := h.paymentRepo.GetByMidtransOrderID(notif.OrderID)
	if err != nil {
		utils.ErrorResponse(w, "Payment not found", http.StatusNotFound)
		return
	}
	
	// Get order
	order, err := h.orderRepo.GetByOrderNumber(notif.OrderID)
	if err != nil {
		utils.ErrorResponse(w, "Order not found", http.StatusNotFound)
		return
	}
	
	// Update payment record
	payment.TransactionID = &notif.TransactionID
	payment.TransactionStatus = &notif.TransactionStatus
	payment.PaymentType = &notif.PaymentType
	payment.StatusMessage = &notif.StatusMessage
	payment.RawResponse = json.RawMessage(body)
	
	if transTime, err := notif.GetTransactionTime(); err == nil {
		payment.TransactionTime = transTime
	}
	
	payment.VANumber = utils.StringPtr(notif.GetVANumber())
	payment.Bank = utils.StringPtr(notif.GetBank())
	payment.FraudStatus = &notif.FraudStatus
	
	if err := h.paymentRepo.Update(payment); err != nil {
		utils.ErrorResponse(w, "Failed to update payment", http.StatusInternalServerError)
		return
	}
	
	// Update order payment status based on transaction status
	switch notif.TransactionStatus {
	case "capture", "settlement":
		// Payment successful
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusPaid, &notif.PaymentType)
		go h.notifService.NotifyPaymentReceived(payment, order)
		
	case "pending":
		// Payment pending
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusPending, &notif.PaymentType)
		
	case "deny", "cancel", "expire":
		// Payment failed
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusFailed, nil)
		
	case "refund":
		// Payment refunded
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusRefunded, nil)
	}
	
	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// internal/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"lendral3n/ordering-system/internal/services/auth"
	"lendral3n/ordering-system/internal/utils"
	"strings"
)

func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.ErrorResponse(w, "Authorization header required", http.StatusUnauthorized)
				return
			}
			
			// Extract token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.ErrorResponse(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
			
			token := parts[1]
			
			// Validate token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				utils.ErrorResponse(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}
			
			// Add user info to context
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "role", claims.Role)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("role").(string)
			
			allowed := false
			for _, role := range roles {
				if userRole == role {
					allowed = true
					break
				}
			}
			
			if !allowed {
				utils.ErrorResponse(w, "Insufficient permissions", http.StatusForbidden)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/cors.go
package middleware

import (
	"net/http"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")
			
			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/logger.go
package middleware

import (
	"net/http"
	"time"

	"lendral3n/ordering-system/pkg/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			
			next.ServeHTTP(wrapped, r)
			
			log.Info(
				"HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", time.Since(start),
				"ip", r.RemoteAddr,
			)
		})
	}
}

// internal/middleware/recovery.go
package middleware

import (
	"net/http"
	"lendral3n/ordering-system/internal/utils"
	"runtime/debug"
)

func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the error and stack trace
					debug.PrintStack()
					
					// Return error response
					utils.ErrorResponse(w, "Internal server error", http.StatusInternalServerError)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/rate_limiter.go
package middleware

import (
	"net/http"
	"lendral3n/ordering-system/internal/utils"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func RateLimiter(rps int) func(http.Handler) http.Handler {
	// Clean up old entries periodically
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			
			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				limiter := rate.NewLimiter(rate.Limit(rps), rps)
				visitors[ip] = &visitor{limiter, time.Now()}
				v = visitors[ip]
			}
			v.lastSeen = time.Now()
			mu.Unlock()
			
			if !v.limiter.Allow() {
				utils.ErrorResponse(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/request_id.go
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			w.Header().Set("X-Request-ID", requestID)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// internal/routes/customer.go
package routes

import (
	"lendral3n/ordering-system/internal/config"
	"lendral3n/ordering-system/internal/handlers/customer"
	"lendral3n/ordering-system/internal/middleware"

	"github.com/gorilla/mux"
)

func SetupCustomerRoutes(router *mux.Router, handlers *customer.Handlers, cfg *config.Config) {
	// Public routes
	api := router.PathPrefix("/api/customer").Subrouter()
	
	// Session routes
	api.HandleFunc("/session/start", handlers.Session.StartSession).Methods("POST")
	api.HandleFunc("/session", handlers.Session.GetSession).Methods("GET")
	api.HandleFunc("/session/end", handlers.Session.EndSession).Methods("POST")
	
	// Menu routes (public)
	api.HandleFunc("/menu/categories", handlers.Menu.GetCategories).Methods("GET")
	api.HandleFunc("/menu/items", handlers.Menu.GetMenuItems).Methods("GET")
	api.HandleFunc("/menu/items/{id}", handlers.Menu.GetMenuItem).Methods("GET")
	api.HandleFunc("/menu/items/{id}/360", handlers.Menu.GetMenu360View).Methods("GET")
	
	// Protected routes (require session)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.RequireSession())
	
	// Order routes
	protected.HandleFunc("/orders", handlers.Order.CreateOrder).Methods("POST")
	protected.HandleFunc("/orders/{id}", handlers.Order.GetOrder).Methods("GET")
	protected.HandleFunc("/orders/session", handlers.Order.GetOrdersBySession).Methods("GET")
	protected.HandleFunc("/assistance", handlers.Order.RequestAssistance).Methods("POST")
	
	// Payment routes
	protected.HandleFunc("/payment/create", handlers.Payment.CreatePayment).Methods("POST")
	protected.HandleFunc("/payment/{order_id}/status", handlers.Payment.GetPaymentStatus).Methods("GET")
	
	// Payment callback (public)
	api.HandleFunc("/payment/finish", handlers.Payment.HandlePaymentFinish).Methods("GET", "POST")
}

// internal/routes/staff.go
package routes

import (
	"lendral3n/ordering-system/internal/config"
	"lendral3n/ordering-system/internal/handlers/staff"
	"lendral3n/ordering-system/internal/middleware"
	"lendral3n/ordering-system/internal/services/auth"

	"github.com/gorilla/mux"
)

func SetupStaffRoutes(router *mux.Router, handlers *staff.Handlers, cfg *config.Config) {
	api := router.PathPrefix("/api/staff").Subrouter()
	
	// Auth service for middleware
	authService := auth.NewService(cfg.JWTSecret, cfg.JWTExpiry)
	
	// Public routes
	api.HandleFunc("/auth/login", handlers.Auth.Login).Methods("POST")
	
	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware(authService))
	
	// Auth routes
	protected.HandleFunc("/auth/profile", handlers.Auth.GetProfile).Methods("GET")
	protected.HandleFunc("/auth/change-password", handlers.Auth.ChangePassword).Methods("POST")
	
	// Order routes
	protected.HandleFunc("/orders", handlers.Order.GetOrders).Methods("GET")
	protected.HandleFunc("/orders/{id}", handlers.Order.GetOrder).Methods("GET")
	protected.HandleFunc("/orders/{id}/status", handlers.Order.UpdateOrderStatus).Methods("PUT")
	protected.HandleFunc("/orders/items/{item_id}/status", handlers.Order.UpdateOrderItemStatus).Methods("PUT")
	protected.HandleFunc("/orders/{id}/invoice", handlers.Order.GenerateInvoice).Methods("GET")
	
	// Menu routes
	protected.HandleFunc("/menu/categories", handlers.Menu.GetCategories).Methods("GET")
	protected.HandleFunc("/menu/categories", handlers.Menu.CreateCategory).Methods("POST")
	protected.HandleFunc("/menu/categories/{id}", handlers.Menu.UpdateCategory).Methods("PUT")
	
	protected.HandleFunc("/menu/items", handlers.Menu.GetMenuItems).Methods("GET")
	protected.HandleFunc("/menu/items", handlers.Menu.CreateMenuItem).Methods("POST")
	protected.HandleFunc("/menu/items/{id}", handlers.Menu.UpdateMenuItem).Methods("PUT")
	protected.HandleFunc("/menu/items/{id}", handlers.Menu.DeleteMenuItem).Methods("DELETE")
	protected.HandleFunc("/menu/items/{id}/media", handlers.Menu.UploadMedia).Methods("POST")
	protected.HandleFunc("/menu/items/{id}/stock", handlers.Menu.UpdateStock).Methods("PUT")
	
	// Payment routes
	protected.HandleFunc("/payments", handlers.Payment.GetPayments).Methods("GET")
	protected.HandleFunc("/payments/{id}/verify", handlers.Payment.VerifyPayment).Methods("PUT")
	
	// Notification routes
	protected.HandleFunc("/notifications", handlers.Notification.GetNotifications).Methods("GET")
	protected.HandleFunc("/notifications/unread-count", handlers.Notification.GetUnreadCount).Methods("GET")
	protected.HandleFunc("/notifications/{id}/read", handlers.Notification.MarkAsRead).Methods("PUT")
	protected.HandleFunc("/notifications/read-all", handlers.Notification.MarkAllAsRead).Methods("PUT")
	
	// Analytics routes
	protected.HandleFunc("/analytics/sales", handlers.Analytics.GetSalesAnalytics).Methods("GET")
	
	// Admin only routes
	adminOnly := protected.PathPrefix("").Subrouter()
	adminOnly.Use(middleware.RequireRole("admin"))
	
	// Admin specific endpoints can be added here
}

// internal/routes/webhook.go
package routes

import (
	"lendral3n/ordering-system/internal/handlers/webhook"

	"github.com/gorilla/mux"
)

func SetupWebhookRoutes(router *mux.Router, handlers *webhook.Handlers) {
	api := router.PathPrefix("/api/webhook").Subrouter()
	
	// Midtrans webhook
	api.HandleFunc("/midtrans", handlers.Midtrans.HandleNotification).Methods("POST")
}

// internal/routes/websocket.go
package routes

import (
	"lendral3n/ordering-system/internal/services/notification"

	"github.com/gorilla/mux"
)

func SetupWebSocketRoutes(router *mux.Router, hub *notification.Hub) {
	router.HandleFunc("/ws", hub.ServeWS)
}

// internal/utils/response.go
package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	
	json.NewEncoder(w).Encode(response)
}

func ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := Response{
		Success: false,
		Error:   message,
	}
	
	json.NewEncoder(w).Encode(response)
}

// internal/utils/validator.go
package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}// internal/handlers/webhook/midtrans.go
package webhook

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/notification"
	"lendral3n/ordering-system/internal/services/payment"
	"lendral3n/ordering-system/internal/utils"
)

type Handlers struct {
	Midtrans *MidtransWebhookHandler
}

func NewHandlers(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	paymentService *payment.MidtransService,
	notifService *notification.Service,
) *Handlers {
	return &Handlers{
		Midtrans: NewMidtransWebhookHandler(paymentRepo, orderRepo, paymentService, notifService),
	}
}

type MidtransWebhookHandler struct {
	paymentRepo    repository.PaymentRepository
	orderRepo      repository.OrderRepository
	paymentService *payment.MidtransService
	notifService   *notification.Service
}

func NewMidtransWebhookHandler(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	paymentService *payment.MidtransService,
	notifService *notification.Service,
) *MidtransWebhookHandler {
	return &MidtransWebhookHandler{
		paymentRepo:    paymentRepo,
		orderRepo:      orderRepo,
		paymentService: paymentService,
		notifService:   notifService,
	}
}

func (h *MidtransWebhookHandler) HandleNotification(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	
	// Parse notification
	notif, err := h.paymentService.ParseNotification(body)
	if err != nil {
		utils.ErrorResponse(w, "Invalid notification", http.StatusBadRequest)
		return
	}
	
	// Get payment record
	payment, err := h.paymentRepo.GetByMidtransOrderID(notif.OrderID)
	if err != nil {
		utils.ErrorResponse(w, "Payment not found", http.StatusNotFound)
		return
	}
	
	// Get order
	order, err := h.orderRepo.GetByOrderNumber(notif.OrderID)
	if err != nil {
		utils.ErrorResponse(w, "Order not found", http.StatusNotFound)
		return
	}
	
	// Update payment record
	payment.TransactionID = &notif.TransactionID
	payment.TransactionStatus = &notif.TransactionStatus
	payment.PaymentType = &notif.PaymentType
	payment.StatusMessage = &notif.StatusMessage
	payment.RawResponse = json.RawMessage(body)
	
	if transTime, err := notif.GetTransactionTime(); err == nil {
		payment.TransactionTime = transTime
	}
	
	payment.VANumber = utils.StringPtr(notif.GetVANumber())
	payment.Bank = utils.StringPtr(notif.GetBank())
	payment.FraudStatus = &notif.FraudStatus
	
	if err := h.paymentRepo.Update(payment); err != nil {
		utils.ErrorResponse(w, "Failed to update payment", http.StatusInternalServerError)
		return
	}
	
	// Update order payment status based on transaction status
	switch notif.TransactionStatus {
	case "capture", "settlement":
		// Payment successful
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusPaid, &notif.PaymentType)
		go h.notifService.NotifyPaymentReceived(payment, order)
		
	case "pending":
		// Payment pending
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusPending, &notif.PaymentType)
		
	case "deny", "cancel", "expire":
		// Payment failed
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusFailed, nil)
		
	case "refund":
		// Payment refunded
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusRefunded, nil)
	}
	
	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// internal/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"lendral3n/ordering-system/internal/services/auth"
	"lendral3n/ordering-system/internal/utils"
	"strings"
)

func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.ErrorResponse(w, "Authorization header required", http.StatusUnauthorized)
				return
			}
			
			// Extract token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.ErrorResponse(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
			
			token := parts[1]
			
			// Validate token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				utils.ErrorResponse(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}
			
			// Add user info to context
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "role", claims.Role)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("role").(string)
			
			allowed := false
			for _, role := range roles {
				if userRole == role {
					allowed = true
					break
				}
			}
			
			if !allowed {
				utils.ErrorResponse(w, "Insufficient permissions", http.StatusForbidden)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/cors.go
package middleware

import (
	"net/http"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")
			
			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/logger.go
package middleware

import (
	"net/http"
	"time"

	"lendral3n/ordering-system/pkg/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			
			next.ServeHTTP(wrapped, r)
			
			log.Info(
				"HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", time.Since(start),
				"ip", r.RemoteAddr,
			)
		})
	}
}

// internal/middleware/recovery.go
package middleware

import (
	"net/http"
	"lendral3n/ordering-system/internal/utils"
	"runtime/debug"
)

func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the error and stack trace
					debug.PrintStack()
					
					// Return error response
					utils.ErrorResponse(w, "Internal server error", http.StatusInternalServerError)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/rate_limiter.go
package middleware

import (
	"net/http"
	"lendral3n/ordering-system/internal/utils"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func RateLimiter(rps int) func(http.Handler) http.Handler {
	// Clean up old entries periodically
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			
			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				limiter := rate.NewLimiter(rate.Limit(rps), rps)
				visitors[ip] = &visitor{limiter, time.Now()}
				v = visitors[ip]
			}
			v.lastSeen = time.Now()
			mu.Unlock()
			
			if !v.limiter.Allow() {
				utils.ErrorResponse(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/request_id.go
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			w.Header().Set("X-Request-ID", requestID)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// internal/routes/customer.go
package routes

import (
	"lendral3n/ordering-system/internal/config"
	"lendral3n/ordering-system/internal/handlers/customer"
	"lendral3n/ordering-system/internal/middleware"

	"github.com/gorilla/mux"
)

func SetupCustomerRoutes(router *mux.Router, handlers *customer.Handlers, cfg *config.Config) {
	// Public routes
	api := router.PathPrefix("/api/customer").Subrouter()
	
	// Session routes
	api.HandleFunc("/session/start", handlers.Session.StartSession).Methods("POST")
	api.HandleFunc("/session", handlers.Session.GetSession).Methods("GET")
	api.HandleFunc("/session/end", handlers.Session.EndSession).Methods("POST")
	
	// Menu routes (public)
	api.HandleFunc("/menu/categories", handlers.Menu.GetCategories).Methods("GET")
	api.HandleFunc("/menu/items", handlers.Menu.GetMenuItems).Methods("GET")
	api.HandleFunc("/menu/items/{id}", handlers.Menu.GetMenuItem).Methods("GET")
	api.HandleFunc("/menu/items/{id}/360", handlers.Menu.GetMenu360View).Methods("GET")
	
	// Protected routes (require session)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.RequireSession())
	
)
	phoneRegex = regexp.MustCompile(`^[+]?[(]?[0-9]{3}[)]?[-\s\.]?[(]?[0-9]{3}[)]?[-\s\.]?[0-9]{4,6}// internal/handlers/webhook/midtrans.go
package webhook

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/notification"
	"lendral3n/ordering-system/internal/services/payment"
	"lendral3n/ordering-system/internal/utils"
)

type Handlers struct {
	Midtrans *MidtransWebhookHandler
}

func NewHandlers(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	paymentService *payment.MidtransService,
	notifService *notification.Service,
) *Handlers {
	return &Handlers{
		Midtrans: NewMidtransWebhookHandler(paymentRepo, orderRepo, paymentService, notifService),
	}
}

type MidtransWebhookHandler struct {
	paymentRepo    repository.PaymentRepository
	orderRepo      repository.OrderRepository
	paymentService *payment.MidtransService
	notifService   *notification.Service
}

func NewMidtransWebhookHandler(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	paymentService *payment.MidtransService,
	notifService *notification.Service,
) *MidtransWebhookHandler {
	return &MidtransWebhookHandler{
		paymentRepo:    paymentRepo,
		orderRepo:      orderRepo,
		paymentService: paymentService,
		notifService:   notifService,
	}
}

func (h *MidtransWebhookHandler) HandleNotification(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	
	// Parse notification
	notif, err := h.paymentService.ParseNotification(body)
	if err != nil {
		utils.ErrorResponse(w, "Invalid notification", http.StatusBadRequest)
		return
	}
	
	// Get payment record
	payment, err := h.paymentRepo.GetByMidtransOrderID(notif.OrderID)
	if err != nil {
		utils.ErrorResponse(w, "Payment not found", http.StatusNotFound)
		return
	}
	
	// Get order
	order, err := h.orderRepo.GetByOrderNumber(notif.OrderID)
	if err != nil {
		utils.ErrorResponse(w, "Order not found", http.StatusNotFound)
		return
	}
	
	// Update payment record
	payment.TransactionID = &notif.TransactionID
	payment.TransactionStatus = &notif.TransactionStatus
	payment.PaymentType = &notif.PaymentType
	payment.StatusMessage = &notif.StatusMessage
	payment.RawResponse = json.RawMessage(body)
	
	if transTime, err := notif.GetTransactionTime(); err == nil {
		payment.TransactionTime = transTime
	}
	
	payment.VANumber = utils.StringPtr(notif.GetVANumber())
	payment.Bank = utils.StringPtr(notif.GetBank())
	payment.FraudStatus = &notif.FraudStatus
	
	if err := h.paymentRepo.Update(payment); err != nil {
		utils.ErrorResponse(w, "Failed to update payment", http.StatusInternalServerError)
		return
	}
	
	// Update order payment status based on transaction status
	switch notif.TransactionStatus {
	case "capture", "settlement":
		// Payment successful
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusPaid, &notif.PaymentType)
		go h.notifService.NotifyPaymentReceived(payment, order)
		
	case "pending":
		// Payment pending
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusPending, &notif.PaymentType)
		
	case "deny", "cancel", "expire":
		// Payment failed
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusFailed, nil)
		
	case "refund":
		// Payment refunded
		h.orderRepo.UpdatePaymentStatus(order.ID, models.PaymentStatusRefunded, nil)
	}
	
	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// internal/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"lendral3n/ordering-system/internal/services/auth"
	"lendral3n/ordering-system/internal/utils"
	"strings"
)

func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.ErrorResponse(w, "Authorization header required", http.StatusUnauthorized)
				return
			}
			
			// Extract token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.ErrorResponse(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
			
			token := parts[1]
			
			// Validate token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				utils.ErrorResponse(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}
			
			// Add user info to context
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "role", claims.Role)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("role").(string)
			
			allowed := false
			for _, role := range roles {
				if userRole == role {
					allowed = true
					break
				}
			}
			
			if !allowed {
				utils.ErrorResponse(w, "Insufficient permissions", http.StatusForbidden)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/cors.go
package middleware

import (
	"net/http"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")
			
			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/logger.go
package middleware

import (
	"net/http"
	"time"

	"lendral3n/ordering-system/pkg/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			
			next.ServeHTTP(wrapped, r)
			
			log.Info(
				"HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", time.Since(start),
				"ip", r.RemoteAddr,
			)
		})
	}
}

// internal/middleware/recovery.go
package middleware

import (
	"net/http"
	"lendral3n/ordering-system/internal/utils"
	"runtime/debug"
)

func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the error and stack trace
					debug.PrintStack()
					
					// Return error response
					utils.ErrorResponse(w, "Internal server error", http.StatusInternalServerError)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/rate_limiter.go
package middleware

import (
	"net/http"
	"lendral3n/ordering-system/internal/utils"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func RateLimiter(rps int) func(http.Handler) http.Handler {
	// Clean up old entries periodically
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			
			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				limiter := rate.NewLimiter(rate.Limit(rps), rps)
				visitors[ip] = &visitor{limiter, time.Now()}
				v = visitors[ip]
			}
			v.lastSeen = time.Now()
			mu.Unlock()
			
			if !v.limiter.Allow() {
				utils.ErrorResponse(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// internal/middleware/request_id.go
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			w.Header().Set("X-Request-ID", requestID)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// internal/routes/customer.go
package routes

import (
	"lendral3n/ordering-system/internal/config"
	"lendral3n/ordering-system/internal/handlers/customer"
	"lendral3n/ordering-system/internal/middleware"

	"github.com/gorilla/mux"
)

func SetupCustomerRoutes(router *mux.Router, handlers *customer.Handlers, cfg *config.Config) {
	// Public routes
	api := router.PathPrefix("/api/customer").Subrouter()
	
	// Session routes
	api.HandleFunc("/session/start", handlers.Session.StartSession).Methods("POST")
	api.HandleFunc("/session", handlers.Session.GetSession).Methods("GET")
	api.HandleFunc("/session/end", handlers.Session.EndSession).Methods("POST")
	
	// Menu routes (public)
	api.HandleFunc("/menu/categories", handlers.Menu.GetCategories).Methods("GET")
	api.HandleFunc("/menu/items", handlers.Menu.GetMenuItems).Methods("GET")
	api.HandleFunc("/menu/items/{id}", handlers.Menu.GetMenuItem).Methods("GET")
	api.HandleFunc("/menu/items/{id}/360", handlers.Menu.GetMenu360View).Methods("GET")
	
	// Protected routes (require session)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.RequireSession())
	
)
)

func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func ValidatePhone(phone string) error {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	
	if !phoneRegex.MatchString(cleaned) {
		return fmt.Errorf("invalid phone format")
	}
	return nil
}

func ValidateRequired(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

func ValidateMinLength(value string, minLength int, fieldName string) error {
	if len(value) < minLength {
		return fmt.Errorf("%s must be at least %d characters", fieldName, minLength)
	}
	return nil
}

func ValidateMaxLength(value string, maxLength int, fieldName string) error {
	if len(value) > maxLength {
		return fmt.Errorf("%s must not exceed %d characters", fieldName, maxLength)
	}
	return nil
}

func ValidatePositiveNumber(value float64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

// internal/utils/helpers.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func GenerateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func GenerateOrderNumber() string {
	timestamp := time.Now().Format("20060102150405")
	random := GenerateRandomString(4)
	return fmt.Sprintf("ORD-%s-%s", timestamp, strings.ToUpper(random))
}

func GenerateInvoiceNumber() string {
	timestamp := time.Now().Format("20060102")
	random := GenerateRandomString(4)
	return fmt.Sprintf("INV-%s-%s", timestamp, strings.ToUpper(random))
}

func FormatCurrency(amount float64) string {
	return fmt.Sprintf("Rp %.0f", amount)
}

func StringPtr(s string) *string {
	return &s
}

func IntPtr(i int) *int {
	return &i
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

// internal/middleware/session.go
package middleware

import (
	"net/http"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/utils"
)

func RequireSession() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Session-Token")
			if token == "" {
				utils.ErrorResponse(w, "Session token required", http.StatusUnauthorized)
				return
			}
			
			// Note: In real implementation, you would need to pass the session repository
			// This is a simplified version
			next.ServeHTTP(w, r)
		})
	}
}

// pkg/logger/logger.go
package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
}

type logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	warnLogger  *log.Logger
}

func New() Logger {
	return &logger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		warnLogger:  log.New(os.Stdout, "[WARN] ", log.LstdFlags),
	}
}

func (l *logger) Info(msg string, keysAndValues ...interface{}) {
	l.log(l.infoLogger, msg, keysAndValues...)
}

func (l *logger) Error(msg string, keysAndValues ...interface{}) {
	l.log(l.errorLogger, msg, keysAndValues...)
}

func (l *logger) Debug(msg string, keysAndValues ...interface{}) {
	l.log(l.debugLogger, msg, keysAndValues...)
}

func (l *logger) Warn(msg string, keysAndValues ...interface{}) {
	l.log(l.warnLogger, msg, keysAndValues...)
}

func (l *logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.log(l.errorLogger, msg, keysAndValues...)
	os.Exit(1)
}

func (l *logger) log(logger *log.Logger, msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) == 0 {
		logger.Println(msg)
		return
	}
	
	fields := ""
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		if i > 0 {
			fields += ", "
		}
		fields += fmt.Sprintf("%s=%v", keysAndValues[i], keysAndValues[i+1])
	}
	
	logger.Printf("%s [%s]", msg, fields)
}

// pkg/errors/errors.go
package errors

import (
	"fmt"
)

type AppError struct {
	Code    string
	Message string
	Details interface{}
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func NewWithDetails(code, message string, details interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common errors
var (
	ErrNotFound          = New("NOT_FOUND", "Resource not found")
	ErrUnauthorized      = New("UNAUTHORIZED", "Unauthorized")
	ErrForbidden         = New("FORBIDDEN", "Forbidden")
	ErrBadRequest        = New("BAD_REQUEST", "Bad request")
	ErrInternalServer    = New("INTERNAL_SERVER", "Internal server error")
	ErrDatabaseError     = New("DATABASE_ERROR", "Database error")
	ErrValidationFailed  = New("VALIDATION_FAILED", "Validation failed")
	ErrDuplicateEntry    = New("DUPLICATE_ENTRY", "Duplicate entry")
	ErrInsufficientStock = New("INSUFFICIENT_STOCK", "Insufficient stock")
	ErrPaymentFailed     = New("PAYMENT_FAILED", "Payment failed")
	ErrSessionExpired    = New("SESSION_EXPIRED", "Session expired")
)