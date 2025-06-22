// internal/services/invoice/generator.go
package invoice

import (
	"bytes"
	"fmt"
	"html/template"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/google/uuid"
)

type Service struct {
	orderRepo   repository.OrderRepository
	paymentRepo repository.PaymentRepository
}

func NewService(orderRepo repository.OrderRepository, paymentRepo repository.PaymentRepository) *Service {
	return &Service{
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
	}
}

type InvoiceData struct {
	InvoiceNumber string
	InvoiceDate   string
	Order         *models.Order
	Payment       *models.Payment
	Items         []InvoiceItem
	Subtotal      float64
	Tax           float64
	ServiceCharge float64
	Total         float64
}

type InvoiceItem struct {
	Name      string
	Quantity  int
	UnitPrice float64
	Total     float64
}

func (s *Service) GenerateInvoice(orderID int) ([]byte, string, error) {
	// Get order with items
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get order: %w", err)
	}
	
	// Get payment
	payment, err := s.paymentRepo.GetByOrderID(orderID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get payment: %w", err)
	}
	
	// Generate invoice number
	invoiceNumber := s.generateInvoiceNumber()
	
	// Prepare invoice data
	invoiceData := InvoiceData{
		InvoiceNumber: invoiceNumber,
		InvoiceDate:   time.Now().Format("02 January 2006"),
		Order:         order,
		Payment:       payment,
		Items:         make([]InvoiceItem, 0, len(order.OrderItems)),
		Subtotal:      order.TotalAmount,
		Tax:           order.TaxAmount,
		ServiceCharge: order.ServiceCharge,
		Total:         order.GrandTotal,
	}
	
	for _, item := range order.OrderItems {
		invoiceData.Items = append(invoiceData.Items, InvoiceItem{
			Name:      item.MenuItem.Name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Total:     item.Subtotal,
		})
	}
	
	// Generate HTML
	html, err := s.generateHTML(invoiceData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate HTML: %w", err)
	}
	
	// Convert to PDF
	pdf, err := s.generatePDF(html)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate PDF: %w", err)
	}
	
	return pdf, invoiceNumber, nil
}

func (s *Service) generateHTML(data InvoiceData) (string, error) {
	tmplStr := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { text-align: center; margin-bottom: 30px; }
        .invoice-info { margin-bottom: 20px; }
        .invoice-info div { margin: 5px 0; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f8f9fa; font-weight: bold; }
        .text-right { text-align: right; }
        .totals { margin-top: 20px; }
        .totals table { width: 300px; margin-left: auto; }
        .totals td { border: none; padding: 8px; }
        .grand-total { font-size: 1.2em; font-weight: bold; }
        .footer { margin-top: 40px; text-align: center; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>INVOICE</h1>
        <h2>Restaurant Name</h2>
        <p>Jl. Example Street No. 123<br>Jakarta, Indonesia</p>
    </div>
    
    <div class="invoice-info">
        <div><strong>Invoice Number:</strong> {{.InvoiceNumber}}</div>
        <div><strong>Date:</strong> {{.InvoiceDate}}</div>
        <div><strong>Order Number:</strong> {{.Order.OrderNumber}}</div>
        <div><strong>Table:</strong> {{.Order.Table.TableNumber}}</div>
        <div><strong>Payment Method:</strong> {{.Payment.PaymentType}}</div>
    </div>
    
    <table>
        <thead>
            <tr>
                <th>Item</th>
                <th class="text-right">Qty</th>
                <th class="text-right">Unit Price</th>
                <th class="text-right">Total</th>
            </tr>
        </thead>
        <tbody>
            {{range .Items}}
            <tr>
                <td>{{.Name}}</td>
                <td class="text-right">{{.Quantity}}</td>
                <td class="text-right">Rp {{printf "%.0f" .UnitPrice}}</td>
                <td class="text-right">Rp {{printf "%.0f" .Total}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
    
    <div class="totals">
        <table>
            <tr>
                <td>Subtotal:</td>
                <td class="text-right">Rp {{printf "%.0f" .Subtotal}}</td>
            </tr>
            <tr>
                <td>Tax (10%):</td>
                <td class="text-right">Rp {{printf "%.0f" .Tax}}</td>
            </tr>
            <tr>
                <td>Service Charge (5%):</td>
                <td class="text-right">Rp {{printf "%.0f" .ServiceCharge}}</td>
            </tr>
            <tr class="grand-total">
                <td>Total:</td>
                <td class="text-right">Rp {{printf "%.0f" .Total}}</td>
            </tr>
        </table>
    </div>
    
    <div class="footer">
        <p>Thank you for dining with us!</p>
        <p>Please visit us again</p>
    </div>
</body>
</html>
`
	
	tmpl, err := template.New("invoice").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

func (s *Service) generatePDF(html string) ([]byte, error) {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, err
	}
	
	// Set global options
	pdfg.Dpi.Set(300)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	
	// Create page from HTML
	page := wkhtmltopdf.NewPageReader(bytes.NewReader([]byte(html)))
	pdfg.AddPage(page)
	
	// Create PDF
	if err := pdfg.Create(); err != nil {
		return nil, err
	}
	
	return pdfg.Bytes(), nil
}

func (s *Service) generateInvoiceNumber() string {
	// Format: INV-YYYYMMDD-XXXX
	now := time.Now()
	dateStr := now.Format("20060102")
	uniqueID := uuid.New().String()[:4]
	return fmt.Sprintf("INV-%s-%s", dateStr, uniqueID)
}
