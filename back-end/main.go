package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"lendral3n/ordering-system/internal/config"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/handler"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/routes"
	"lendral3n/ordering-system/internal/services/media"
	"lendral3n/ordering-system/internal/services/notification"
	"lendral3n/ordering-system/internal/services/payment"
	"lendral3n/ordering-system/internal/services/qrcode"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
)

func main() {
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

	// Auto migrate - Migration order matters for foreign keys
	migrationModels := []interface{}{
		&models.Table{},
		&models.Staff{},
		// Menu related - category first, then items
		&models.MenuCategory{},
		&models.MenuItem{},
		// Session and orders
		&models.CustomerSession{},
		&models.Order{},
		&models.OrderItem{},
		// Others
		&models.Payment{},
		&models.MediaFile{},
		&models.Notification{},
		&models.InventoryLog{},
	}

	for _, model := range migrationModels {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("Failed to migrate %T: %v", model, err)
		}
	}

	// Initialize services
	cloudinaryService, err := media.NewCloudinaryService(cfg.CloudinaryURL)
	if err != nil {
		log.Fatal("Failed to initialize Cloudinary:", err)
	}

	midtransService := payment.NewMidtransService(cfg)
	qrService := qrcode.NewService(cfg.BaseURL)
	notificationHub := notification.NewHub()

	// Start notification hub
	go notificationHub.Run()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "Internal Server Error"

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
			}

			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   message,
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.AllowedOrigins, ","),
		AllowHeaders:     "Origin, Content-Type, Accept, X-Session-Token",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// WebSocket upgrade middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Initialize handlers
	h := handlers.NewHandlers(
		db,
		cloudinaryService,
		midtransService,
		qrService,
		notificationHub,
		cfg,
	)

	// Setup routes
	routes.SetupRoutes(app, h)

	// Setup WebSocket
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		notificationHub.HandleWebSocket(c)
	}))

	// Start server
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown
	if err := app.Shutdown(); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}