// internal/services/payment/midtrans.go
package payment

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"lendral3n/ordering-system/internal/config"
	"lendral3n/ordering-system/internal/models"
	"strings"
	"time"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type MidtransService struct {
	snapClient snap.Client
	coreClient coreapi.Client
	serverKey  string
}

func NewMidtransService(cfg *config.Config) *MidtransService {
	snapClient := snap.Client{}
	coreClient := coreapi.Client{}

	var env midtrans.EnvironmentType
	if cfg.MidtransEnv == "production" {
		env = midtrans.Production
	} else {
		env = midtrans.Sandbox
	}

	snapClient.New(cfg.MidtransServerKey, env)
	coreClient.New(cfg.MidtransServerKey, env)

	return &MidtransService{
		snapClient: snapClient,
		coreClient: coreClient,
		serverKey:  cfg.MidtransServerKey,
	}
}

type CreateTransactionRequest struct {
	Order         *models.Order
	CustomerName  string
	CustomerPhone string
	CustomerEmail string
}

type TransactionResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

func (s *MidtransService) CreateTransaction(req CreateTransactionRequest) (*TransactionResponse, error) {
	// Prepare item details
	items := make([]midtrans.ItemDetails, 0, len(req.Order.OrderItems)+2)

	for _, item := range req.Order.OrderItems {
		items = append(items, midtrans.ItemDetails{
			ID:    fmt.Sprintf("ITEM-%d", item.MenuItemID),
			Name:  item.MenuItem.Name,
			Price: int64(item.UnitPrice),
			Qty:   int32(item.Quantity),
		})
	}

	// Add tax
	if req.Order.TaxAmount > 0 {
		items = append(items, midtrans.ItemDetails{
			ID:    "TAX",
			Name:  "Tax 10%",
			Price: int64(req.Order.TaxAmount),
			Qty:   1,
		})
	}

	// Add service charge
	if req.Order.ServiceCharge > 0 {
		items = append(items, midtrans.ItemDetails{
			ID:    "SERVICE",
			Name:  "Service Charge 5%",
			Price: int64(req.Order.ServiceCharge),
			Qty:   1,
		})
	}

	// Create customer details
	custDetail := &midtrans.CustomerDetails{
		FName: req.CustomerName,
		Phone: req.CustomerPhone,
	}

	if req.CustomerEmail != "" {
		custDetail.Email = req.CustomerEmail
	}

	// Create transaction request
	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.Order.OrderNumber,
			GrossAmt: int64(req.Order.GrandTotal),
		},
		ItemDetails:     &items,
		CustomerDetail:  custDetail,
		EnabledPayments: s.getEnabledPayments(),
		Expiry: &snap.ExpiryDetails{
			Duration: 1,
			Unit:     "hour",
		},
	}

	// Create transaction
	snapResp, err := s.snapClient.CreateTransaction(snapReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create midtrans transaction: %w", err)
	}

	return &TransactionResponse{
		Token:       snapResp.Token,
		RedirectURL: snapResp.RedirectURL,
	}, nil
}

func (s *MidtransService) GetTransactionStatus(orderID string) (*coreapi.TransactionStatusResponse, error) {
	resp, err := s.coreClient.CheckTransaction(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check transaction status: %w", err)
	}
	return resp, nil
}

func (s *MidtransService) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	// Create signature
	input := orderID + statusCode + grossAmount + s.serverKey
	hash := sha512.New()
	hash.Write([]byte(input))
	expectedSignature := hex.EncodeToString(hash.Sum(nil))

	return expectedSignature == signatureKey
}

func (s *MidtransService) ParseNotification(payload []byte) (*NotificationPayload, error) {
	var notif NotificationPayload
	if err := json.Unmarshal(payload, &notif); err != nil {
		return nil, err
	}

	// Verify signature
	if !s.VerifySignature(notif.OrderID, notif.StatusCode, notif.GrossAmount, notif.SignatureKey) {
		return nil, fmt.Errorf("invalid signature")
	}

	return &notif, nil
}

func (s *MidtransService) getEnabledPayments() []snap.SnapPaymentType {
	return []snap.SnapPaymentType{
		snap.PaymentTypeBankTransfer,
		snap.PaymentTypeCreditCard,
		snap.PaymentTypeGopay,
		snap.PaymentTypeQris,
		snap.PaymentTypeShopeepay,
		snap.PaymentTypeBCAKlikpay,
		snap.PaymentTypeAlfamart,
		snap.PaymentTypeIndomaret,
	}
}

type NotificationPayload struct {
	TransactionTime   string          `json:"transaction_time"`
	TransactionStatus string          `json:"transaction_status"`
	TransactionID     string          `json:"transaction_id"`
	StatusMessage     string          `json:"status_message"`
	StatusCode        string          `json:"status_code"`
	SignatureKey      string          `json:"signature_key"`
	PaymentType       string          `json:"payment_type"`
	OrderID           string          `json:"order_id"`
	MerchantID        string          `json:"merchant_id"`
	GrossAmount       string          `json:"gross_amount"`
	FraudStatus       string          `json:"fraud_status"`
	Currency          string          `json:"currency"`
	VANumbers         []VANumber      `json:"va_numbers"`
	PaymentAmounts    []PaymentAmount `json:"payment_amounts"`
	PermataVANumber   string          `json:"permata_va_number"`
	BillKey           string          `json:"bill_key"`
	BillerCode        string          `json:"biller_code"`
	Store             string          `json:"store"`
	ApprovalCode      string          `json:"approval_code"`
	Bank              string          `json:"bank"`
}

type VANumber struct {
	Bank     string `json:"bank"`
	VANumber string `json:"va_number"`
}

type PaymentAmount struct {
	PaidAt string `json:"paid_at"`
	Amount string `json:"amount"`
}

func (p *NotificationPayload) GetVANumber() string {
	if p.PermataVANumber != "" {
		return p.PermataVANumber
	}

	if len(p.VANumbers) > 0 {
		return p.VANumbers[0].VANumber
	}

	return ""
}

func (p *NotificationPayload) GetBank() string {
	if p.Bank != "" {
		return p.Bank
	}

	if len(p.VANumbers) > 0 {
		return p.VANumbers[0].Bank
	}

	if strings.Contains(p.PaymentType, "bank_transfer") && p.PermataVANumber != "" {
		return "permata"
	}

	return ""
}

func (p *NotificationPayload) GetTransactionTime() (*time.Time, error) {
	if p.TransactionTime == "" {
		return nil, nil
	}

	// Midtrans time format: "2024-01-15 10:30:45"
	t, err := time.Parse("2006-01-02 15:04:05", p.TransactionTime)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
