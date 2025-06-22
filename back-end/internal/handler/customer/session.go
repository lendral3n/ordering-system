// internal/handlers/customer/session.go
package customer

import (
	"encoding/json"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/utils"
	"net/http"
)

type SessionHandler struct {
	sessionRepo repository.SessionRepository
}

func NewSessionHandler(sessionRepo repository.SessionRepository) *SessionHandler {
	return &SessionHandler{
		sessionRepo: sessionRepo,
	}
}

type StartSessionRequest struct {
	TableNumber   string `json:"table_number"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
}

type StartSessionResponse struct {
	SessionToken string        `json:"session_token"`
	Table        *models.Table `json:"table"`
}

func (h *SessionHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	var req StartSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.TableNumber == "" {
		utils.ErrorResponse(w, "Table number is required", http.StatusBadRequest)
		return
	}

	// Get table by number
	tableRepo := repository.NewTableRepository(h.sessionRepo.(*repository.sessionRepository).db)
	table, err := tableRepo.GetByTableNumber(req.TableNumber)
	if err != nil {
		utils.ErrorResponse(w, "Invalid table number", http.StatusNotFound)
		return
	}

	// Check if table is available
	if table.Status != models.TableStatusAvailable {
		utils.ErrorResponse(w, "Table is not available", http.StatusConflict)
		return
	}

	// Check for existing active session
	existingSession, _ := h.sessionRepo.GetByTableID(table.ID, true)
	if existingSession != nil {
		utils.ErrorResponse(w, "Table already has an active session", http.StatusConflict)
		return
	}

	// Create new session
	session := &models.CustomerSession{
		TableID:       table.ID,
		CustomerName:  &req.CustomerName,
		CustomerPhone: &req.CustomerPhone,
	}

	if err := h.sessionRepo.Create(session); err != nil {
		utils.ErrorResponse(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Update table status
	tableRepo.UpdateStatus(table.ID, models.TableStatusOccupied)

	response := StartSessionResponse{
		SessionToken: session.SessionToken,
		Table:        table,
	}

	utils.SuccessResponse(w, "Session started successfully", response)
}

func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Session-Token")
	if token == "" {
		utils.ErrorResponse(w, "Session token required", http.StatusUnauthorized)
		return
	}

	session, err := h.sessionRepo.GetByToken(token)
	if err != nil {
		utils.ErrorResponse(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	if session.EndedAt != nil {
		utils.ErrorResponse(w, "Session has ended", http.StatusUnauthorized)
		return
	}

	utils.SuccessResponse(w, "Session retrieved", session)
}

func (h *SessionHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Session-Token")
	if token == "" {
		utils.ErrorResponse(w, "Session token required", http.StatusUnauthorized)
		return
	}

	session, err := h.sessionRepo.GetByToken(token)
	if err != nil {
		utils.ErrorResponse(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// End session
	if err := h.sessionRepo.EndSession(token); err != nil {
		utils.ErrorResponse(w, "Failed to end session", http.StatusInternalServerError)
		return
	}

	// Update table status to available
	tableRepo := repository.NewTableRepository(h.sessionRepo.(*repository.sessionRepository).db)
	tableRepo.UpdateStatus(session.TableID, models.TableStatusAvailable)

	utils.SuccessResponse(w, "Session ended successfully", nil)
}
