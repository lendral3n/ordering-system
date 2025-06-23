package repository

import (
	"crypto/rand"
	"encoding/hex"
	"lendral3n/ordering-system/internal/models"
	"time"
	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(session *models.CustomerSession) error {
	// Generate session token if not provided
	if session.SessionToken == "" {
		session.SessionToken = r.generateSessionToken()
	}
	
	return r.db.Create(session).Error
}

func (r *sessionRepository) GetByToken(token string) (*models.CustomerSession, error) {
	var session models.CustomerSession
	err := r.db.Preload("Table").
		Where("session_token = ?", token).
		First(&session).Error
	
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetByTableID(tableID uint, activeOnly bool) (*models.CustomerSession, error) {
	var session models.CustomerSession
	query := r.db.Where("table_id = ?", tableID)
	
	if activeOnly {
		query = query.Where("ended_at IS NULL")
	}
	
	err := query.Order("started_at DESC").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) EndSession(token string) error {
	result := r.db.Model(&models.CustomerSession{}).
		Where("session_token = ? AND ended_at IS NULL", token).
		Update("ended_at", gorm.Expr("CURRENT_TIMESTAMP"))
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	
	return nil
}

func (r *sessionRepository) CleanupExpiredSessions(expiryDuration time.Duration) error {
	expiryTime := time.Now().Add(-expiryDuration)
	
	return r.db.Model(&models.CustomerSession{}).
		Where("ended_at IS NULL AND started_at < ?", expiryTime).
		Update("ended_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *sessionRepository) generateSessionToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based token
		return "sess_" + hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(bytes)
}