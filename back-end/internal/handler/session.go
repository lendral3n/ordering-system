package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"lendral3n/ordering-system/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type StartSessionRequest struct {
	TableNumber   string `json:"table_number" validate:"required"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
}

type StartSessionResponse struct {
	SessionToken string        `json:"session_token"`
	Table        *models.Table `json:"table"`
}

func (h *Handlers) StartSession(c *fiber.Ctx) error {
	var req StartSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Get table by number
	var table models.Table
	if err := h.DB.Where("table_number = ?", req.TableNumber).First(&table).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Table not found",
		})
	}

	// Check if table is available
	if table.Status != models.TableStatusAvailable {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"error":   "Table is not available",
		})
	}

	// Check for existing active session
	var existingSession models.CustomerSession
	err := h.DB.Where("table_id = ? AND ended_at IS NULL", table.ID).First(&existingSession).Error
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"error":   "Table already has an active session",
		})
	}

	// Generate session token
	sessionToken := generateSessionToken()

	// Create new session
	session := models.CustomerSession{
		SessionToken:  sessionToken,
		TableID:       table.ID,
		CustomerName:  &req.CustomerName,
		CustomerPhone: &req.CustomerPhone,
	}

	if err := h.DB.Create(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create session",
		})
	}

	// Update table status
	table.Status = models.TableStatusOccupied
	h.DB.Save(&table)

	// Load table data for response
	h.DB.Preload("Table").First(&session, session.ID)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Session started successfully",
		"data": StartSessionResponse{
			SessionToken: session.SessionToken,
			Table:        &session.Table,
		},
	})
}

func (h *Handlers) GetSession(c *fiber.Ctx) error {
	token := c.Get("X-Session-Token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Session token required",
		})
	}

	var session models.CustomerSession
	if err := h.DB.Preload("Table").Where("session_token = ?", token).First(&session).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid session",
		})
	}

	if session.EndedAt != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Session has ended",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Session retrieved",
		"data":    session,
	})
}

func (h *Handlers) EndSession(c *fiber.Ctx) error {
	token := c.Get("X-Session-Token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Session token required",
		})
	}

	var session models.CustomerSession
	if err := h.DB.Where("session_token = ?", token).First(&session).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid session",
		})
	}

	// End session
	h.DB.Model(&session).Update("ended_at", gorm.Expr("CURRENT_TIMESTAMP"))

	// Update table status to available
	h.DB.Model(&models.Table{}).Where("id = ?", session.TableID).Update("status", models.TableStatusAvailable)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Session ended successfully",
	})
}

func generateSessionToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
