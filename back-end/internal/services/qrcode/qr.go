package qrcode

import (
	"encoding/base64"
	"fmt"
	"net/url"

	qr "github.com/skip2/go-qrcode"
)

type QRService struct {
	baseURL string
}

func NewService(baseURL string) *QRService {
	return &QRService{
		baseURL: baseURL,
	}
}

func (s *QRService) GenerateTableQRCode(tableNumber string) ([]byte, error) {
	// Create URL for scanning
	scanURL := fmt.Sprintf("%s/scan?table=%s", s.baseURL, url.QueryEscape(tableNumber))

	// Generate QR code
	qrCode, err := qr.Encode(scanURL, qr.Medium, 512)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return qrCode, nil
}

func (s *QRService) GenerateTableQRCodeBase64(tableNumber string) (string, error) {
	qrCode, err := s.GenerateTableQRCode(tableNumber)
	if err != nil {
		return "", err
	}

	// Convert to base64
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(qrCode), nil
}