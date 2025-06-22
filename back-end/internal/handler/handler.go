package handlers

import (
	"lendral3n/ordering-system/internal/config"
	"lendral3n/ordering-system/internal/services/media"
	"lendral3n/ordering-system/internal/services/notification"
	"lendral3n/ordering-system/internal/services/payment"
	"lendral3n/ordering-system/internal/services/qrcode"

	"gorm.io/gorm"
)

type Handlers struct {
	DB                *gorm.DB
	CloudinaryService *media.CloudinaryService
	MidtransService   *payment.MidtransService
	QRService         *qrcode.QRService
	NotificationHub   *notification.Hub
	Config            *config.Config
}

func NewHandlers(
	db *gorm.DB,
	cloudinaryService *media.CloudinaryService,
	midtransService *payment.MidtransService,
	qrService *qrcode.QRService,
	notificationHub *notification.Hub,
	config *config.Config,
) *Handlers {
	return &Handlers{
		DB:                db,
		CloudinaryService: cloudinaryService,
		MidtransService:   midtransService,
		QRService:         qrService,
		NotificationHub:   notificationHub,
		Config:            config,
	}
}