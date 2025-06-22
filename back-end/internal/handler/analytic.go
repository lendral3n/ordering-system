package handlers

import (
	"fmt"
	"lendral3n/ordering-system/internal/models"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) GetSalesAnalytics(c *fiber.Ctx) error {
	// Parse date range
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		// Default to last 30 days
		end := time.Now()
		start := end.AddDate(0, 0, -30)
		startDate = start.Format("2006-01-02")
		endDate = end.Format("2006-01-02")
	}

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	end = end.Add(24 * time.Hour) // Include the entire end date

	// Get orders within date range
	var orders []models.Order
	if err := h.DB.Preload("OrderItems.MenuItem").
		Where("created_at >= ? AND created_at < ?", start, end).
		Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get analytics data",
		})
	}

	// Calculate analytics
	analytics := h.calculateAnalytics(orders)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Analytics retrieved",
		"data":    analytics,
	})
}

func (h *Handlers) calculateAnalytics(orders []models.Order) map[string]interface{} {
	totalRevenue := 0.0
	totalOrders := len(orders)
	completedOrders := 0
	cancelledOrders := 0

	itemsSold := make(map[string]int)
	hourlyOrders := make(map[int]int)
	dailyRevenue := make(map[string]float64)
	categoryRevenue := make(map[string]float64)

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

			// Popular items and category revenue
			for _, item := range order.OrderItems {
				if item.MenuItem.Name != "" {
					itemsSold[item.MenuItem.Name] += item.Quantity
					if item.MenuItem.Category.Name != "" {
						categoryRevenue[item.MenuItem.Category.Name] += item.Subtotal
					}
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

	// Calculate completion rate
	completionRate := 0.0
	if totalOrders > 0 {
		completionRate = float64(completedOrders) / float64(totalOrders) * 100
	}

	// Get peak hours
	peakHours := h.getPeakHours(hourlyOrders)

	// Format daily revenue for chart
	dailyRevenueArray := h.formatDailyRevenue(dailyRevenue)

	// Format category revenue
	categoryRevenueArray := h.formatCategoryRevenue(categoryRevenue)

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_revenue":       totalRevenue,
			"total_orders":        totalOrders,
			"completed_orders":    completedOrders,
			"cancelled_orders":    cancelledOrders,
			"average_order_value": avgOrderValue,
			"completion_rate":     completionRate,
		},
		"top_selling_items":    topItems,
		"hourly_distribution":  hourlyOrders,
		"daily_revenue":        dailyRevenueArray,
		"category_revenue":     categoryRevenueArray,
		"peak_hours":           peakHours,
	}
}

func (h *Handlers) getTopItems(items map[string]int, limit int) []map[string]interface{} {
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

func (h *Handlers) getPeakHours(hourlyOrders map[int]int) []map[string]interface{} {
	type hourData struct {
		Hour   int
		Orders int
	}

	var hours []hourData
	for hour, orders := range hourlyOrders {
		hours = append(hours, hourData{hour, orders})
	}

	// Sort by orders descending
	sort.Slice(hours, func(i, j int) bool {
		return hours[i].Orders > hours[j].Orders
	})

	// Take top 3 peak hours
	result := make([]map[string]interface{}, 0)
	for i := 0; i < 3 && i < len(hours); i++ {
		result = append(result, map[string]interface{}{
			"hour":   fmt.Sprintf("%02d:00", hours[i].Hour),
			"orders": hours[i].Orders,
		})
	}

	return result
}

func (h *Handlers) formatDailyRevenue(dailyRevenue map[string]float64) []map[string]interface{} {
	type dailyData struct {
		Date    string
		Revenue float64
	}

	var days []dailyData
	for date, revenue := range dailyRevenue {
		days = append(days, dailyData{date, revenue})
	}

	// Sort by date
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date < days[j].Date
	})

	result := make([]map[string]interface{}, 0)
	for _, day := range days {
		result = append(result, map[string]interface{}{
			"date":    day.Date,
			"revenue": day.Revenue,
		})
	}

	return result
}

func (h *Handlers) formatCategoryRevenue(categoryRevenue map[string]float64) []map[string]interface{} {
	type categoryData struct {
		Category string
		Revenue  float64
	}

	var categories []categoryData
	for category, revenue := range categoryRevenue {
		categories = append(categories, categoryData{category, revenue})
	}

	// Sort by revenue descending
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Revenue > categories[j].Revenue
	})

	result := make([]map[string]interface{}, 0)
	for _, cat := range categories {
		result = append(result, map[string]interface{}{
			"category": cat.Category,
			"revenue":  cat.Revenue,
		})
	}

	return result
}

// Additional analytics endpoints

func (h *Handlers) GetTableAnalytics(c *fiber.Ctx) error {
	// Get table usage statistics
	var tableStats []struct {
		TableID     uint
		TableNumber string
		OrderCount  int64
		Revenue     float64
	}

	h.DB.Table("orders").
		Select("orders.table_id, tables.table_number, COUNT(orders.id) as order_count, SUM(orders.grand_total) as revenue").
		Joins("JOIN tables ON tables.id = orders.table_id").
		Where("orders.status = ?", models.OrderStatusCompleted).
		Group("orders.table_id, tables.table_number").
		Order("revenue DESC").
		Scan(&tableStats)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Table analytics retrieved",
		"data":    tableStats,
	})
}

func (h *Handlers) GetMenuPerformance(c *fiber.Ctx) error {
	// Get menu item performance
	var itemStats []struct {
		MenuItemID   uint
		ItemName     string
		CategoryName string
		OrderCount   int64
		Quantity     int64
		Revenue      float64
	}

	h.DB.Table("order_items").
		Select(`
			order_items.menu_item_id,
			menu_items.name as item_name,
			menu_categories.name as category_name,
			COUNT(DISTINCT order_items.order_id) as order_count,
			SUM(order_items.quantity) as quantity,
			SUM(order_items.subtotal) as revenue
		`).
		Joins("JOIN menu_items ON menu_items.id = order_items.menu_item_id").
		Joins("JOIN menu_categories ON menu_categories.id = menu_items.category_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.status = ?", models.OrderStatusCompleted).
		Group("order_items.menu_item_id, menu_items.name, menu_categories.name").
		Order("revenue DESC").
		Scan(&itemStats)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Menu performance retrieved",
		"data":    itemStats,
	})
}