// internal/repository/staff_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type staffRepository struct {
	db *database.DB
}

func NewStaffRepository(db *database.DB) StaffRepository {
	return &staffRepository{db: db}
}

func (r *staffRepository) GetByID(id int) (*models.Staff, error) {
	query := `
		SELECT id, username, email, password_hash, full_name,
			   role, is_active, last_login, created_at, updated_at
		FROM staff
		WHERE id = $1
	`

	var staff models.Staff
	err := r.db.QueryRow(query, id).Scan(
		&staff.ID,
		&staff.Username,
		&staff.Email,
		&staff.PasswordHash,
		&staff.FullName,
		&staff.Role,
		&staff.IsActive,
		&staff.LastLogin,
		&staff.CreatedAt,
		&staff.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("staff not found")
	}

	return &staff, err
}

func (r *staffRepository) GetByUsername(username string) (*models.Staff, error) {
	query := `
		SELECT id, username, email, password_hash, full_name,
			   role, is_active, last_login, created_at, updated_at
		FROM staff
		WHERE username = $1
	`

	var staff models.Staff
	err := r.db.QueryRow(query, username).Scan(
		&staff.ID,
		&staff.Username,
		&staff.Email,
		&staff.PasswordHash,
		&staff.FullName,
		&staff.Role,
		&staff.IsActive,
		&staff.LastLogin,
		&staff.CreatedAt,
		&staff.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("staff not found")
	}

	return &staff, err
}

func (r *staffRepository) GetByEmail(email string) (*models.Staff, error) {
	query := `
		SELECT id, username, email, password_hash, full_name,
			   role, is_active, last_login, created_at, updated_at
		FROM staff
		WHERE email = $1
	`

	var staff models.Staff
	err := r.db.QueryRow(query, email).Scan(
		&staff.ID,
		&staff.Username,
		&staff.Email,
		&staff.PasswordHash,
		&staff.FullName,
		&staff.Role,
		&staff.IsActive,
		&staff.LastLogin,
		&staff.CreatedAt,
		&staff.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("staff not found")
	}

	return &staff, err
}

func (r *staffRepository) Create(staff *models.Staff) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(staff.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO staff (username, email, password_hash, full_name, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		staff.Username,
		staff.Email,
		string(hashedPassword),
		staff.FullName,
		staff.Role,
		staff.IsActive,
	).Scan(&staff.ID, &staff.CreatedAt, &staff.UpdatedAt)
}

func (r *staffRepository) Update(staff *models.Staff) error {
	query := `
		UPDATE staff
		SET email = $2, full_name = $3, role = $4, is_active = $5
		WHERE id = $1
		RETURNING updated_at
	`

	return r.db.QueryRow(
		query,
		staff.ID,
		staff.Email,
		staff.FullName,
		staff.Role,
		staff.IsActive,
	).Scan(&staff.UpdatedAt)
}

func (r *staffRepository) UpdateLastLogin(id int) error {
	query := `UPDATE staff SET last_login = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *staffRepository) GetAll(activeOnly bool) ([]models.Staff, error) {
	query := `
		SELECT id, username, email, password_hash, full_name,
			   role, is_active, last_login, created_at, updated_at
		FROM staff
	`

	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY full_name"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staffList []models.Staff
	for rows.Next() {
		var staff models.Staff
		err := rows.Scan(
			&staff.ID,
			&staff.Username,
			&staff.Email,
			&staff.PasswordHash,
			&staff.FullName,
			&staff.Role,
			&staff.IsActive,
			&staff.LastLogin,
			&staff.CreatedAt,
			&staff.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		staffList = append(staffList, staff)
	}

	return staffList, rows.Err()
}
