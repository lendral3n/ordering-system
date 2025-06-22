// internal/services/notification/websocket.go
package notification

import (
	"encoding/json"
	"fmt"
	"lendral3n/ordering-system/internal/models"
	"lendral3n/ordering-system/internal/repository"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	role    string
	userID  int
	tableID int
}

type Message struct {
	Type    string      `json:"type"`
	Target  string      `json:"target"` // "all", "staff", "table:{id}", "user:{id}"
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected: role=%s, userID=%d, tableID=%d", client.role, client.userID, client.tableID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.mu.Unlock()
				log.Printf("Client disconnected: role=%s, userID=%d", client.role, client.userID)
			} else {
				h.mu.Unlock()
			}

		case message := <-h.broadcast:
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				if h.shouldSendToClient(client, msg) {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) shouldSendToClient(client *Client, msg Message) bool {
	switch msg.Target {
	case "all":
		return true
	case "staff":
		return client.role == "staff"
	case "customer":
		return client.role == "customer"
	default:
		// Handle specific targets like "table:1" or "user:5"
		if len(msg.Target) > 6 && msg.Target[:6] == "table:" {
			tableID := 0
			fmt.Sscanf(msg.Target, "table:%d", &tableID)
			return client.tableID == tableID
		}
		if len(msg.Target) > 5 && msg.Target[:5] == "user:" {
			userID := 0
			fmt.Sscanf(msg.Target, "user:%d", &userID)
			return client.userID == userID
		}
	}
	return false
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Get client info from query params
	role := r.URL.Query().Get("role")
	userID := 0
	tableID := 0

	if role == "staff" {
		fmt.Sscanf(r.URL.Query().Get("user_id"), "%d", &userID)
	} else if role == "customer" {
		fmt.Sscanf(r.URL.Query().Get("table_id"), "%d", &tableID)
	}

	client := &Client{
		hub:     h,
		conn:    conn,
		send:    make(chan []byte, 256),
		role:    role,
		userID:  userID,
		tableID: tableID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Service for notifications
type Service struct {
	hub       *Hub
	notifRepo repository.NotificationRepository
}

func NewService(notifRepo repository.NotificationRepository) *Service {
	return &Service{
		notifRepo: notifRepo,
	}
}

func (s *Service) SetHub(hub *Hub) {
	s.hub = hub
}

func (s *Service) NotifyNewOrder(order *models.Order) error {
	// Create notification in database
	notif := &models.StaffNotification{
		OrderID: &order.ID,
		Type:    models.NotificationNewOrder,
		Message: fmt.Sprintf("New order #%s from table %s", order.OrderNumber, order.Table.TableNumber),
	}

	if err := s.notifRepo.Create(notif); err != nil {
		return err
	}

	// Send WebSocket notification
	msg := Message{
		Type:    "new_order",
		Target:  "staff",
		Message: notif.Message,
		Data: map[string]interface{}{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
			"table_number": order.Table.TableNumber,
			"total":        order.GrandTotal,
		},
	}

	return s.broadcast(msg)
}

func (s *Service) NotifyPaymentReceived(payment *models.Payment, order *models.Order) error {
	// Create notification in database
	notif := &models.StaffNotification{
		OrderID: &order.ID,
		Type:    models.NotificationPaymentReceived,
		Message: fmt.Sprintf("Payment received for order #%s", order.OrderNumber),
	}

	if err := s.notifRepo.Create(notif); err != nil {
		return err
	}

	// Notify staff
	staffMsg := Message{
		Type:    "payment_received",
		Target:  "staff",
		Message: notif.Message,
		Data: map[string]interface{}{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
			"amount":       payment.GrossAmount,
			"method":       payment.PaymentType,
		},
	}

	if err := s.broadcast(staffMsg); err != nil {
		return err
	}

	// Notify customer
	customerMsg := Message{
		Type:    "payment_confirmed",
		Target:  fmt.Sprintf("table:%d", order.TableID),
		Message: "Payment confirmed! Thank you.",
		Data: map[string]interface{}{
			"order_id": order.ID,
			"status":   "paid",
		},
	}

	return s.broadcast(customerMsg)
}

func (s *Service) NotifyOrderStatusUpdate(order *models.Order) error {
	// Notify customer about status change
	msg := Message{
		Type:    "order_status_updated",
		Target:  fmt.Sprintf("table:%d", order.TableID),
		Message: fmt.Sprintf("Order #%s status: %s", order.OrderNumber, order.Status),
		Data: map[string]interface{}{
			"order_id": order.ID,
			"status":   order.Status,
		},
	}

	// Special notification for ready orders
	if order.Status == models.OrderStatusReady {
		// Create staff notification
		notif := &models.StaffNotification{
			OrderID: &order.ID,
			Type:    models.NotificationOrderReady,
			Message: fmt.Sprintf("Order #%s is ready to serve", order.OrderNumber),
		}

		if err := s.notifRepo.Create(notif); err != nil {
			return err
		}

		// Send special ready notification
		msg.Type = "order_ready"
		msg.Message = "Your order is ready!"
	}

	return s.broadcast(msg)
}

func (s *Service) NotifyAssistanceRequest(tableID int, tableNumber string) error {
	// Create notification
	notif := &models.StaffNotification{
		Type:    models.NotificationAssistanceRequest,
		Message: fmt.Sprintf("Table %s needs assistance", tableNumber),
	}

	if err := s.notifRepo.Create(notif); err != nil {
		return err
	}

	// Send WebSocket notification to staff
	msg := Message{
		Type:    "assistance_request",
		Target:  "staff",
		Message: notif.Message,
		Data: map[string]interface{}{
			"table_id":     tableID,
			"table_number": tableNumber,
		},
	}

	return s.broadcast(msg)
}

func (s *Service) broadcast(msg Message) error {
	if s.hub == nil {
		return fmt.Errorf("WebSocket hub not initialized")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	s.hub.broadcast <- data
	return nil
}
