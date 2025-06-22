
// internal/repository/payment_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"lendral3n/ordering-system/internal/database"
	"lendral3n/ordering-system/internal/models"
)

type paymentRepository struct {
	db *database.DB
}

func NewPaymentRepository(db *database.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *models.Payment) error {
	query := `
		INSERT INTO payments (
			order_id, midtrans_order_id, midtrans_transaction_id,
			payment_type, transaction_status, transaction_time,
			gross_amount, currency, va_number, bank, fraud_status,
			signature_key, status_message, raw_response
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`
	
	return r.db.QueryRow(
		query,
		payment.OrderID,
		payment.MidtransOrderID,
		payment.MidtransTransactionID,
		payment.PaymentType,
		payment.TransactionStatus,
		payment.TransactionTime,
		payment.GrossAmount,
		payment.Currency,
		payment.VANumber,
		payment.Bank,
		payment.FraudStatus,
		payment.SignatureKey,
		payment.StatusMessage,
		payment.RawResponse,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)
}

func (r *paymentRepository) GetByOrderID(orderID int) (*models.Payment, error) {
	query := `
		SELECT 
			id, order_id, midtrans_order_id, midtrans_transaction_id,
			payment_type, transaction_status, transaction_time,
			gross_amount, currency, va_number, bank, fraud_status,
			signature_key, status_message, raw_response, created_at, updated_at
		FROM payments
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	
	var payment models.Payment
	err := r.db.QueryRow(query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.MidtransOrderID,
		&payment.MidtransTransactionID,
		&payment.PaymentType,
		&payment.TransactionStatus,
		&payment.TransactionTime,
		&payment.GrossAmount,
		&payment.Currency,
		&payment.VANumber,
		&payment.Bank,
		&payment.FraudStatus,
		&payment.SignatureKey,
		&payment.StatusMessage,
		&payment.RawResponse,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment not found")
	}
	
	return &payment, err
}

func (r *paymentRepository) GetByMidtransOrderID(midtransOrderID string) (*models.Payment, error) {
	query := `
		SELECT 
			id, order_id, midtrans_order_id, midtrans_transaction_id,
			payment_type, transaction_status, transaction_time,
			gross_amount, currency, va_number, bank, fraud_status,
			signature_key, status_message, raw_response, created_at, updated_at
		FROM payments
		WHERE midtrans_order_id = $1
	`
	
	var payment models.Payment
	err := r.db.QueryRow(query, midtransOrderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.MidtransOrderID,
		&payment.MidtransTransactionID,
		&payment.PaymentType,
		&payment.TransactionStatus,
		&payment.TransactionTime,
		&payment.GrossAmount,
		&payment.Currency,
		&payment.VANumber,
		&payment.Bank,
		&payment.FraudStatus,
		&payment.SignatureKey,
		&payment.StatusMessage,
		&payment.RawResponse,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment not found")
	}
	
	return &payment, err
}

func (r *paymentRepository) Update(payment *models.Payment) error {
	query := `
		UPDATE payments
		SET midtrans_transaction_id = $2, payment_type = $3,
			transaction_status = $4, transaction_time = $5,
			va_number = $6, bank = $7, fraud_status = $8,
			signature_key = $9, status_message = $10, raw_response = $11
		WHERE id = $1
		RETURNING updated_at
	`
	
	return r.db.QueryRow(
		query,
		payment.ID,
		payment.MidtransTransactionID,
		payment.PaymentType,
		payment.TransactionStatus,
		payment.TransactionTime,
		payment.VANumber,
		payment.Bank,
		payment.FraudStatus,
		payment.SignatureKey,
		payment.StatusMessage,
		payment.RawResponse,
	).Scan(&payment.UpdatedAt)
}

func (r *paymentRepository) GetAll(filter PaymentFilter) ([]models.Payment, error) {
	query := `
		SELECT 
			p.id, p.order_id, p.midtrans_order_id, p.midtrans_transaction_id,
			p.payment_type, p.transaction_status, p.transaction_time,
			p.gross_amount, p.currency, p.va_number, p.bank, p.fraud_status,
			p.signature_key, p.status_message, p.raw_response, p.created_at, p.updated_at
		FROM payments p
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 0
	
	if filter.Status != nil {
		argCount++
		query += fmt.Sprintf(" AND p.transaction_status = $%d", argCount)
		args = append(args, *filter.Status)
	}
	
	if filter.DateFrom != nil {
		argCount++
		query += fmt.Sprintf(" AND p.created_at >= $%d", argCount)
		args = append(args, *filter.DateFrom)
	}
	
	if filter.DateTo != nil {
		argCount++
		query += fmt.Sprintf(" AND p.created_at <= $%d", argCount)
		args = append(args, *filter.DateTo)
	}
	
	query += " ORDER BY p.created_at DESC"
	
	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}
	
	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var payments []models.Payment
	for rows.Next() {
		var payment models.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.OrderID,
			&payment.MidtransOrderID,
			&payment.MidtransTransactionID,
			&payment.PaymentType,
			&payment.TransactionStatus,
			&payment.TransactionTime,
			&payment.GrossAmount,
			&payment.Currency,
			&payment.VANumber,
			&payment.Bank,
			&payment.FraudStatus,
			&payment.SignatureKey,
			&payment.StatusMessage,
			&payment.RawResponse,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	
	return payments, rows.Err()
}
