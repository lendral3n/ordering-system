package notification

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
	"lendral3n/ordering-system/internal/models"
)

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
	tableID int
}

type Message struct {
	Type    string      `json:"type"`
	Target  string      `json:"target"` // "all", "staff", "table:{id}"
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
			log.Printf("Client connected: role=%s, tableID=%d", client.role, client.tableID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.mu.Unlock()
				log.Printf("Client disconnected: role=%s", client.role)
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
		// Handle specific targets like "table:1"
		if len(msg.Target) > 6 && msg.Target[:6] == "table:" {
			tableID := 0
			fmt.Sscanf(msg.Target, "table:%d", &tableID)
			return client.tableID == tableID
		}
	}
	return false
}

func (h *Hub) HandleWebSocket(c *websocket.Conn) {
	// Get client info from query params
	role := c.Query("role", "customer")
	tableID := c.Query("table_id", "0")

	tableIDInt := 0
	fmt.Sscanf(tableID, "%d", &tableIDInt)

	client := &Client{
		hub:     h,
		conn:    c,
		send:    make(chan []byte, 256),
		role:    role,
		tableID: tableIDInt,
	}

	client.hub.register <- client

	// Start goroutines
	go client.writePump()
	client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

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
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

// Broadcast methods
func (h *Hub) BroadcastNewOrder(order *models.Order) {
	msg := Message{
		Type:    "new_order",
		Target:  "staff",
		Message: fmt.Sprintf("New order #%s from table %s", order.OrderNumber, order.Table.TableNumber),
		Data: map[string]interface{}{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
			"table_number": order.Table.TableNumber,
			"total":        order.GrandTotal,
		},
	}
	h.broadcast <- mustMarshal(msg)
}

func (h *Hub) BroadcastPaymentReceived(payment *models.Payment, order *models.Order) {
	// Notify staff
	staffMsg := Message{
		Type:    "payment_received",
		Target:  "staff",
		Message: fmt.Sprintf("Payment received for order #%s", order.OrderNumber),
		Data: map[string]interface{}{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
			"amount":       payment.GrossAmount,
			"method":       payment.PaymentType,
		},
	}
	h.broadcast <- mustMarshal(staffMsg)

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
	h.broadcast <- mustMarshal(customerMsg)
}

func (h *Hub) BroadcastOrderStatusUpdate(order *models.Order) {
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
		msg.Type = "order_ready"
		msg.Message = "Your order is ready!"
	}

	h.broadcast <- mustMarshal(msg)
}

func (h *Hub) BroadcastAssistanceRequest(tableID uint, tableNumber string) {
	msg := Message{
		Type:    "assistance_request",
		Target:  "staff",
		Message: fmt.Sprintf("Table %s needs assistance", tableNumber),
		Data: map[string]interface{}{
			"table_id":     tableID,
			"table_number": tableNumber,
		},
	}
	h.broadcast <- mustMarshal(msg)
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return []byte{}
	}
	return data
}