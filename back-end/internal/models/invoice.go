
// internal/models/invoice.go
package models

import (
	"time"
)

type Invoice struct {
	ID               int        `json:"id" db:"id"`
	InvoiceNumber    string     `json:"invoice_number" db:"invoice_number"`
	OrderID          int        `json:"order_id" db:"order_id"`
	IssuedDate       time.Time  `json:"issued_date" db:"issued_date"`
	DueDate          *time.Time `json:"due_date" db:"due_date"`
	PDFURL           *string    `json:"pdf_url" db:"pdf_url"`
	SentToEmail      *string    `json:"sent_to_email" db:"sent_to_email"`
	SentToWhatsApp   *string    `json:"sent_to_whatsapp" db:"sent_to_whatsapp"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	
	// Relations
	Order *Order `json:"order,omitempty"`
}