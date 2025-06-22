// scripts/seed.go
package main

import (
	"database/sql"
	"fmt"
	"lendral3n/ordering-system/internal/config"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/auth"
	"lendral3n/ordering-system/internal/services/qrcode"
	"log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize services
	authService := auth.NewService(cfg.JWTSecret, cfg.JWTExpiry)
	qrService := qrcode.NewService(cfg.BaseURL)

	// Initialize repositories
	staffRepo := repository.NewStaffRepository(db)
	tableRepo := repository.NewTableRepository(db)
	menuRepo := repository.NewMenuRepository(db)

	// Seed staff users
	if err := seedStaff(staffRepo, authService); err != nil {
		log.Fatal("Failed to seed staff:", err)
	}

	// Seed tables with QR codes
	if err := seedTables(tableRepo, qrService); err != nil {
		log.Fatal("Failed to seed tables:", err)
	}

	// Seed menu categories and items
	if err := seedMenu(menuRepo); err != nil {
		log.Fatal("Failed to seed menu:", err)
	}

	log.Println("Database seeding completed successfully!")
}

func seedStaff(repo repository.StaffRepository, authService *auth.Service) error {
	log.Println("Seeding staff users...")

	staffMembers := []struct {
		Username string
		Email    string
		Password string
		FullName string
		Role     models.StaffRole
	}{
		{
			Username: "admin",
			Email:    "admin@restaurant.com",
			Password: "admin123",
			FullName: "System Administrator",
			Role:     models.StaffRoleAdmin,
		},
		{
			Username: "cashier1",
			Email:    "cashier1@restaurant.com",
			Password: "cashier123",
			FullName: "John Cashier",
			Role:     models.StaffRoleCashier,
		},
		{
			Username: "waiter1",
			Email:    "waiter1@restaurant.com",
			Password: "waiter123",
			FullName: "Jane Waiter",
			Role:     models.StaffRoleWaiter,
		},
		{
			Username: "kitchen1",
			Email:    "kitchen1@restaurant.com",
			Password: "kitchen123",
			FullName: "Chef Gordon",
			Role:     models.StaffRoleKitchen,
		},
	}

	for _, s := range staffMembers {
		// Check if user already exists
		existing, _ := repo.GetByUsername(s.Username)
		if existing != nil {
			log.Printf("Staff %s already exists, skipping...", s.Username)
			continue
		}

		// Hash password
		hashedPassword, err := authService.HashPassword(s.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", s.Username, err)
		}

		staff := &models.Staff{
			Username:     s.Username,
			Email:        s.Email,
			PasswordHash: hashedPassword,
			FullName:     s.FullName,
			Role:         s.Role,
			IsActive:     true,
		}

		if err := repo.Create(staff); err != nil {
			return fmt.Errorf("failed to create staff %s: %w", s.Username, err)
		}

		log.Printf("Created staff: %s (%s)", s.Username, s.Role)
	}

	return nil
}

func seedTables(repo repository.TableRepository, qrService *qrcode.Service) error {
	log.Println("Seeding tables...")

	tables := []struct {
		Number   string
		Capacity int
	}{
		{"T01", 4},
		{"T02", 4},
		{"T03", 6},
		{"T04", 2},
		{"T05", 8},
		{"T06", 4},
		{"T07", 6},
		{"T08", 2},
		{"VIP1", 10},
		{"VIP2", 12},
	}

	for _, t := range tables {
		// Check if table already exists
		existing, _ := repo.GetByTableNumber(t.Number)
		if existing != nil {
			log.Printf("Table %s already exists, skipping...", t.Number)
			continue
		}

		// Generate QR code
		qrCode, err := qrService.GenerateTableQRCodeBase64(t.Number)
		if err != nil {
			return fmt.Errorf("failed to generate QR code for table %s: %w", t.Number, err)
		}

		table := &models.Table{
			TableNumber: t.Number,
			QRCode:      qrCode,
			Status:      models.TableStatusAvailable,
			Capacity:    t.Capacity,
		}

		if err := repo.Create(table); err != nil {
			return fmt.Errorf("failed to create table %s: %w", t.Number, err)
		}

		log.Printf("Created table: %s (capacity: %d)", t.Number, t.Capacity)
	}

	return nil
}

func seedMenu(repo repository.MenuRepository) error {
	log.Println("Seeding menu categories and items...")

	// Seed categories
	categories := []models.MenuCategory{
		{
			Name:         "Appetizers",
			Description:  strPtr("Start your meal with our delicious appetizers"),
			DisplayOrder: 1,
			IsActive:     true,
		},
		{
			Name:         "Soups",
			Description:  strPtr("Warm and comforting soups"),
			DisplayOrder: 2,
			IsActive:     true,
		},
		{
			Name:         "Main Course",
			Description:  strPtr("Our signature main dishes"),
			DisplayOrder: 3,
			IsActive:     true,
		},
		{
			Name:         "Seafood",
			Description:  strPtr("Fresh from the ocean"),
			DisplayOrder: 4,
			IsActive:     true,
		},
		{
			Name:         "Beverages",
			Description:  strPtr("Refreshing drinks and beverages"),
			DisplayOrder: 5,
			IsActive:     true,
		},
		{
			Name:         "Desserts",
			Description:  strPtr("Sweet treats to end your meal"),
			DisplayOrder: 6,
			IsActive:     true,
		},
	}

	categoryIDs := make(map[string]int)

	for _, c := range categories {
		// Check if category exists
		existing, _ := repo.GetCategories(false)
		found := false
		for _, e := range existing {
			if e.Name == c.Name {
				categoryIDs[c.Name] = e.ID
				found = true
				log.Printf("Category %s already exists, skipping...", c.Name)
				break
			}
		}

		if !found {
			if err := repo.CreateCategory(&c); err != nil {
				return fmt.Errorf("failed to create category %s: %w", c.Name, err)
			}
			categoryIDs[c.Name] = c.ID
			log.Printf("Created category: %s", c.Name)
		}
	}

	// Seed menu items
	menuItems := []struct {
		Category        string
		Name            string
		Description     string
		Price           float64
		PreparationTime int
		StockQuantity   *int
	}{
		// Appetizers
		{
			Category:        "Appetizers",
			Name:            "Spring Rolls",
			Description:     "Crispy vegetable spring rolls with sweet chili sauce",
			Price:           45000,
			PreparationTime: 10,
		},
		{
			Category:        "Appetizers",
			Name:            "Chicken Satay",
			Description:     "Grilled chicken skewers with peanut sauce",
			Price:           55000,
			PreparationTime: 15,
		},
		{
			Category:        "Appetizers",
			Name:            "Crispy Calamari",
			Description:     "Golden fried squid rings with tartar sauce",
			Price:           65000,
			PreparationTime: 12,
		},
		// Soups
		{
			Category:        "Soups",
			Name:            "Tom Yum Soup",
			Description:     "Spicy Thai soup with shrimp and mushrooms",
			Price:           50000,
			PreparationTime: 15,
		},
		{
			Category:        "Soups",
			Name:            "Mushroom Cream Soup",
			Description:     "Rich and creamy mushroom soup",
			Price:           40000,
			PreparationTime: 10,
		},
		// Main Course
		{
			Category:        "Main Course",
			Name:            "Nasi Goreng Special",
			Description:     "Indonesian fried rice with chicken, shrimp, and fried egg",
			Price:           75000,
			PreparationTime: 20,
		},
		{
			Category:        "Main Course",
			Name:            "Beef Rendang",
			Description:     "Slow-cooked beef in coconut milk and spices",
			Price:           95000,
			PreparationTime: 30,
		},
		{
			Category:        "Main Course",
			Name:            "Grilled Chicken",
			Description:     "Herb-marinated grilled chicken with vegetables",
			Price:           85000,
			PreparationTime: 25,
		},
		// Seafood
		{
			Category:        "Seafood",
			Name:            "Grilled Salmon",
			Description:     "Fresh salmon with lemon butter sauce",
			Price:           120000,
			PreparationTime: 25,
		},
		{
			Category:        "Seafood",
			Name:            "Garlic Butter Prawns",
			Description:     "Jumbo prawns sautéed in garlic butter",
			Price:           110000,
			PreparationTime: 20,
		},
		// Beverages
		{
			Category:        "Beverages",
			Name:            "Fresh Orange Juice",
			Description:     "Freshly squeezed orange juice",
			Price:           25000,
			PreparationTime: 5,
			StockQuantity:   intPtr(50),
		},
		{
			Category:        "Beverages",
			Name:            "Iced Coffee",
			Description:     "Indonesian style iced coffee",
			Price:           30000,
			PreparationTime: 5,
		},
		{
			Category:        "Beverages",
			Name:            "Mineral Water",
			Description:     "600ml bottled water",
			Price:           15000,
			PreparationTime: 1,
			StockQuantity:   intPtr(100),
		},
		// Desserts
		{
			Category:        "Desserts",
			Name:            "Chocolate Lava Cake",
			Description:     "Warm chocolate cake with vanilla ice cream",
			Price:           45000,
			PreparationTime: 15,
		},
		{
			Category:        "Desserts",
			Name:            "Tiramisu",
			Description:     "Classic Italian coffee-flavored dessert",
			Price:           50000,
			PreparationTime: 5,
		},
	}

	for _, item := range menuItems {
		categoryID, ok := categoryIDs[item.Category]
		if !ok {
			log.Printf("Category %s not found, skipping item %s", item.Category, item.Name)
			continue
		}

		// Check if item already exists
		items, _ := repo.GetMenuItems(&categoryID, false)
		found := false
		for _, i := range items {
			if i.Name == item.Name {
				found = true
				log.Printf("Menu item %s already exists, skipping...", item.Name)
				break
			}
		}

		if !found {
			menuItem := &models.MenuItem{
				CategoryID:      categoryID,
				Name:            item.Name,
				Description:     strPtr(item.Description),
				Price:           item.Price,
				IsAvailable:     true,
				PreparationTime: intToNullInt64(item.PreparationTime),
			}

			if item.StockQuantity != nil {
				menuItem.StockQuantity = intToNullInt64(*item.StockQuantity)
			}

			if err := repo.CreateMenuItem(menuItem); err != nil {
				return fmt.Errorf("failed to create menu item %s: %w", item.Name, err)
			}
			log.Printf("Created menu item: %s (Rp %.0f)", item.Name, item.Price)
		}
	}

	return nil
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func intToNullInt64(i int) sql.NullInt64 {
	return sql.NullInt64{
		Int64: int64(i),
		Valid: true,
	}
}
