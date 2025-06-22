
// internal/repository/session_repo.go
package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/models"
	"time"
)

type sessionRepository struct {
	db *database.DB
}

func NewSessionRepository(db *database.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(session *models.CustomerSession) error {
	// Generate secure session token
	session.SessionToken = r.generateSessionToken()
	
	query := `
		INSERT INTO customer_sessions (session_token, table_id, customer_name, customer_phone)
		VALUES ($1, $2, $3, $4)
		RETURNING id, started_at
	`
	
	return r.db.QueryRow(
		query,
		session.SessionToken,
		session.TableID,
		session.CustomerName,
		session.CustomerPhone,
	).Scan(&session.ID, &session.StartedAt)
}

func (r *sessionRepository) GetByToken(token string) (*models.CustomerSession, error) {
	query := `
		SELECT 
			cs.id, cs.session_token, cs.table_id, cs.customer_name,
			cs.customer_phone, cs.started_at, cs.ended_at,
			t.id, t.table_number, t.qr_code, t.status, t.capacity,
			t.created_at, t.updated_at
		FROM customer_sessions cs
		LEFT JOIN tables t ON cs.table_id = t.id
		WHERE cs.session_token = $1
	`
	
	var session models.CustomerSession
	var table models.Table
	
	err := r.db.QueryRow(query, token).Scan(
		&session.ID,
		&session.SessionToken,
		&session.TableID,
		&session.CustomerName,
		&session.CustomerPhone,
		&session.StartedAt,
		&session.EndedAt,
		&table.ID,
		&table.TableNumber,
		&table.QRCode,
		&table.Status,
		&table.Capacity,
		&table.CreatedAt,
		&table.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, err
	}
	
	session.Table = &table
	return &session, nil
}

func (r *sessionRepository) GetByTableID(tableID int, activeOnly bool) (*models.CustomerSession, error) {
	query := `
		SELECT id, session_token, table_id, customer_name,
			   customer_phone, started_at, ended_at
		FROM customer_sessions
		WHERE table_id = $1
	`
	
	if activeOnly {
		query += " AND ended_at IS NULL"
	}
	
	query += " ORDER BY started_at DESC LIMIT 1"
	
	var session models.CustomerSession
	err := r.db.QueryRow(query, tableID).Scan(
		&session.ID,
		&session.SessionToken,
		&session.TableID,
		&session.CustomerName,
		&session.CustomerPhone,
		&session.StartedAt,
		&session.EndedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	
	return &session, err
}

func (r *sessionRepository) EndSession(token string) error {
	query := `
		UPDATE customer_sessions
		SET ended_at = CURRENT_TIMESTAMP
		WHERE session_token = $1 AND ended_at IS NULL
	`
	
	result, err := r.db.Exec(query, token)
	if err != nil {
		return err
	}
	
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if affected == 0 {
		return fmt.Errorf("session not found or already ended")
	}
	
	return nil
}

func (r *sessionRepository) CleanupExpiredSessions(expiryDuration time.Duration) error {
	query := `
		UPDATE customer_sessions
		SET ended_at = CURRENT_TIMESTAMP
		WHERE ended_at IS NULL
		AND started_at < $1
	`
	
	expiryTime := time.Now().Add(-expiryDuration)
	_, err := r.db.Exec(query, expiryTime)
	return err
}

func (r *sessionRepository) generateSessionToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based token
		return fmt.Sprintf("sess_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}