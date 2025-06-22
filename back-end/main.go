package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lendral3n/ordering-system/internal/config"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/handlers/customer"
	"lendral3n/ordering-system/internal/handlers/staff"
	"lendral3n/ordering-system/internal/handlers/webhook"
	"lendral3n/ordering-system/internal/middleware"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/routes"
	"lendral3n/ordering-system/internal/services/auth"
	"lendral3n/ordering-system/internal/services/invoice"
	"lendral3n/ordering-system/internal/services/media"
	"lendral3n/ordering-system/internal/services/notification"
	"lendral3n/ordering-system/internal/services/payment"
	"lendral3n/ordering-system/internal/services/qrcode"
	"lendral3n/ordering-system/pkg/logger"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Initialize logger
	log := logger.New()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db, cfg.MigrationsPath); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize repositories
	repos := initializeRepositories(db)

	// Initialize services
	services := initializeServices(cfg, repos)

	// Initialize handlers
	handlers := initializeHandlers(repos, services)

	// Initialize WebSocket hub
	wsHub := notification.NewHub()
	go wsHub.Run()
	services.NotificationService.SetHub(wsHub)

	// Initialize router
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.Logger(log))
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())

	// Setup routes
	routes.SetupCustomerRoutes(router, handlers.CustomerHandlers, cfg)
	routes.SetupStaffRoutes(router, handlers.StaffHandlers, cfg)
	routes.SetupWebhookRoutes(router, handlers.WebhookHandlers)
	routes.SetupWebSocketRoutes(router, wsHub)

	// Setup static file server for media
	router.PathPrefix("/media/").Handler(
		http.StripPrefix("/media/", http.FileServer(http.Dir(cfg.MediaPath))),
	)

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	handler := c.Handler(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info(fmt.Sprintf("Server starting on port %s", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Info("Server exited")
}

type Repositories struct {
	TableRepo        repository.TableRepository
	MenuRepo         repository.MenuRepository
	OrderRepo        repository.OrderRepository
	PaymentRepo      repository.PaymentRepository
	StaffRepo        repository.StaffRepository
	SessionRepo      repository.SessionRepository
	NotificationRepo repository.NotificationRepository
	MediaRepo        repository.MediaRepository
}

type Services struct {
	AuthService         *auth.Service
	PaymentService      *payment.MidtransService
	QRCodeService       *qrcode.Service
	InvoiceService      *invoice.Service
	NotificationService *notification.Service
	MediaService        *media.Service
}

type Handlers struct {
	CustomerHandlers *customer.Handlers
	StaffHandlers    *staff.Handlers
	WebhookHandlers  *webhook.Handlers
}

func initializeRepositories(db *database.DB) *Repositories {
	return &Repositories{
		TableRepo:        repository.NewTableRepository(db),
		MenuRepo:         repository.NewMenuRepository(db),
		OrderRepo:        repository.NewOrderRepository(db),
		PaymentRepo:      repository.NewPaymentRepository(db),
		StaffRepo:        repository.NewStaffRepository(db),
		SessionRepo:      repository.NewSessionRepository(db),
		NotificationRepo: repository.NewNotificationRepository(db),
		MediaRepo:        repository.NewMediaRepository(db),
	}
}

func initializeServices(cfg *config.Config, repos *Repositories) *Services {
	return &Services{
		AuthService:         auth.NewService(cfg.JWTSecret, cfg.JWTExpiry),
		PaymentService:      payment.NewMidtransService(cfg),
		QRCodeService:       qrcode.NewService(cfg.BaseURL),
		InvoiceService:      invoice.NewService(repos.OrderRepo, repos.PaymentRepo),
		NotificationService: notification.NewService(repos.NotificationRepo),
		MediaService:        media.NewService(cfg.MediaPath, cfg.BaseURL),
	}
}

func initializeHandlers(repos *Repositories, services *Services) *Handlers {
	return &Handlers{
		CustomerHandlers: customer.NewHandlers(
			repos.SessionRepo,
			repos.MenuRepo,
			repos.OrderRepo,
			repos.PaymentRepo,
			services.PaymentService,
			services.NotificationService,
		),
		StaffHandlers: staff.NewHandlers(
			repos.StaffRepo,
			repos.MenuRepo,
			repos.OrderRepo,
			repos.PaymentRepo,
			repos.MediaRepo,
			services.AuthService,
			services.MediaService,
			services.InvoiceService,
			services.NotificationService,
		),
		WebhookHandlers: webhook.NewHandlers(
			repos.PaymentRepo,
			repos.OrderRepo,
			services.PaymentService,
			services.NotificationService,
		),
	}
}
