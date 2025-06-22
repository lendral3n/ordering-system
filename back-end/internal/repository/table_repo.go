
package repository

import (
	"database/sql"
	"fmt"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/models"
)

type tableRepository struct {
	db *database.DB
}

func NewTableRepository(db *database.DB) TableRepository {
	return &tableRepository{db: db}
}

func (r *tableRepository) GetByID(id int) (*models.Table, error) {
	query := `
		SELECT id, table_number, qr_code, status, capacity, created_at, updated_at
		FROM tables
		WHERE id = $1
	`
	
	var table models.Table
	err := r.db.QueryRow(query, id).Scan(
		&table.ID,
		&table.TableNumber,
		&table.QRCode,
		&table.Status,
		&table.Capacity,
		&table.CreatedAt,
		&table.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("table not found")
	}
	
	return &table, err
}

func (r *tableRepository) GetByTableNumber(tableNumber string) (*models.Table, error) {
	query := `
		SELECT id, table_number, qr_code, status, capacity, created_at, updated_at
		FROM tables
		WHERE table_number = $1
	`
	
	var table models.Table
	err := r.db.QueryRow(query, tableNumber).Scan(
		&table.ID,
		&table.TableNumber,
		&table.QRCode,
		&table.Status,
		&table.Capacity,
		&table.CreatedAt,
		&table.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("table not found")
	}
	
	return &table, err
}

func (r *tableRepository) GetAll() ([]models.Table, error) {
	query := `
		SELECT id, table_number, qr_code, status, capacity, created_at, updated_at
		FROM tables
		ORDER BY table_number
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tables []models.Table
	for rows.Next() {
		var table models.Table
		err := rows.Scan(
			&table.ID,
			&table.TableNumber,
			&table.QRCode,
			&table.Status,
			&table.Capacity,
			&table.CreatedAt,
			&table.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	
	return tables, rows.Err()
}

func (r *tableRepository) Create(table *models.Table) error {
	query := `
		INSERT INTO tables (table_number, qr_code, status, capacity)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	
	return r.db.QueryRow(
		query,
		table.TableNumber,
		table.QRCode,
		table.Status,
		table.Capacity,
	).Scan(&table.ID, &table.CreatedAt, &table.UpdatedAt)
}

func (r *tableRepository) Update(table *models.Table) error {
	query := `
		UPDATE tables
		SET table_number = $2, status = $3, capacity = $4
		WHERE id = $1
		RETURNING updated_at
	`
	
	return r.db.QueryRow(
		query,
		table.ID,
		table.TableNumber,
		table.Status,
		table.Capacity,
	).Scan(&table.UpdatedAt)
}

func (r *tableRepository) UpdateStatus(id int, status models.TableStatus) error {
	query := `UPDATE tables SET status = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, status)
	return err
}
