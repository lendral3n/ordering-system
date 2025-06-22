// internal/handlers/customer/handlers.go
package customer

import (
	"lendral3n/ordering-system/internal/repository"
	"lendral3n/ordering-system/internal/services/notification"
	"lendral3n/ordering-system/internal/services/payment"
)

type Handlers struct {
	Session *SessionHandler
	Menu    *MenuHandler
	Order   *OrderHandler
	Payment *PaymentHandler
}

func NewHandlers(
	sessionRepo repository.SessionRepository,
	menuRepo repository.MenuRepository,
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	paymentService *payment.MidtransService,
	notifService *notification.Service,
) *Handlers {
	return &Handlers{
		Session: NewSessionHandler(sessionRepo),
		Menu:    NewMenuHandler(menuRepo),
		Order:   NewOrderHandler(orderRepo, menuRepo, notifService),
		Payment: NewPaymentHandler(orderRepo, paymentRepo, paymentService, notifService),
	}
}
