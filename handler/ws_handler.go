package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	response "hospital/pkg"
	"hospital/schema"
	"hospital/service"
)

// ========================================
// WEBSOCKET UPGRADER
// ========================================

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin cho phep tat ca origin (dev mode).
	// Production nen kiem tra origin cu the.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ========================================
// CLIENT — dai dien 1 ket noi WebSocket
// ========================================

// Client la 1 user dang ket noi WebSocket vao 1 phong chat.
type Client struct {
	conn   *websocket.Conn
	userID uint64
	role   string
	roomID uint64 // conversation_id
	send   chan []byte
}

// ========================================
// HUB — quan ly tat ca phong chat
// ========================================

// Hub quan ly cac phong chat va broadcast tin nhan.
type Hub struct {
	rooms      map[uint64]map[*Client]bool // roomID -> set of clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMsg
	mu         sync.RWMutex
}

// BroadcastMsg chua tin nhan can gui cho phong.
type BroadcastMsg struct {
	RoomID  uint64
	Payload []byte
}

// NewHub tao Hub moi va chay goroutine xu ly.
func NewHub() *Hub {
	h := &Hub{
		rooms:      make(map[uint64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMsg, 256),
	}
	go h.run()
	return h
}

// run vong lap chinh cua Hub, xu ly register/unregister/broadcast.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.roomID] == nil {
				h.rooms[client.roomID] = make(map[*Client]bool)
			}
			h.rooms[client.roomID][client] = true
			h.mu.Unlock()
			log.Printf("[WS] User %d joined room %d", client.userID, client.roomID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.roomID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.rooms, client.roomID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[WS] User %d left room %d", client.userID, client.roomID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.rooms[msg.RoomID]; ok {
				for client := range clients {
					select {
					case client.send <- msg.Payload:
					default:
						// Client bi lag, dong ket noi
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// ========================================
// WS MESSAGE TYPES
// ========================================

// wsIncomingMsg tin nhan client gui len server.
type wsIncomingMsg struct {
	Type        string `json:"type"`         // text, image, voice
	TextContent string `json:"text_content"`
	MediaURL    string `json:"media_url"`
}

// wsOutgoingMsg tin nhan server gui ve client.
type wsOutgoingMsg struct {
	MessageID  uint64            `json:"message_id"`
	SenderID   uint64            `json:"sender_id"`
	SenderType schema.SenderType `json:"sender_type"`
	Type       schema.MessageType `json:"type"`
	TextContent string           `json:"text_content"`
	MediaURL    string           `json:"media_url"`
	CreatedAt   string           `json:"created_at"`
}

// ========================================
// WS HANDLER
// ========================================

// WSHandler xu ly ket noi WebSocket.
type WSHandler struct {
	hub     *Hub
	chatSvc *service.ChatService
}

// NewWSHandler tao WSHandler voi Hub rieng.
func NewWSHandler(chatSvc *service.ChatService) *WSHandler {
	return &WSHandler{
		hub:     NewHub(),
		chatSvc: chatSvc,
	}
}

// HandleWS xu ly upgrade HTTP -> WebSocket.
// URL: GET /api/ws/chat?conversation_id=X&token=Y
func (h *WSHandler) HandleWS(c *gin.Context) {
	// 1. Lay conversation_id tu query
	convIDStr := c.Query("conversation_id")
	if convIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing conversation_id"})
		return
	}
	convID, err := strconv.ParseUint(convIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation_id"})
		return
	}

	// 2. Xac thuc token tu query param (WS khong dung header Authorization)
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	claims, err := response.ParseToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// 3. Upgrade HTTP -> WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] Upgrade failed: %v", err)
		return
	}

	// 4. Tao client va dang ky vao Hub
	client := &Client{
		conn:   conn,
		userID: claims.UserID,
		role:   claims.Role,
		roomID: convID,
		send:   make(chan []byte, 256),
	}

	h.hub.register <- client

	// 5. Chay 2 goroutine doc/ghi
	go h.writePump(client)
	go h.readPump(client)
}

// ========================================
// READ PUMP — doc tin nhan tu client
// ========================================

// readPump doc tin tu WebSocket client -> luu DB -> broadcast.
func (h *WSHandler) readPump(client *Client) {
	defer func() {
		h.hub.unregister <- client
		client.conn.Close()
	}()

	for {
		_, msgBytes, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[WS] Read error user %d: %v", client.userID, err)
			}
			break
		}

		// Parse tin nhan
		var incoming wsIncomingMsg
		if err := json.Unmarshal(msgBytes, &incoming); err != nil {
			log.Printf("[WS] Invalid message from user %d: %v", client.userID, err)
			continue
		}

		// Xac dinh sender_type
		senderType := schema.SenderTypeUser
		if client.role == "admin" || client.role == "coordinator" || client.role == "staff" {
			senderType = schema.SenderTypeStaff
		}

		// Luu vao DB qua ChatService
		msg, err := h.chatSvc.SendMessage(
			client.roomID,
			client.userID,
			senderType,
			schema.MessageType(incoming.Type),
			incoming.TextContent,
			incoming.MediaURL,
		)
		if err != nil {
			log.Printf("[WS] SendMessage failed user %d: %v", client.userID, err)
			continue
		}

		// Tao outgoing message
		outgoing := wsOutgoingMsg{
			MessageID:   msg.MessageID,
			SenderID:    msg.SenderID,
			SenderType:  msg.SenderType,
			Type:        msg.Type,
			TextContent: msg.TextContent,
			MediaURL:    msg.MediaURL,
			CreatedAt:   msg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		payload, err := json.Marshal(outgoing)
		if err != nil {
			log.Printf("[WS] Marshal failed: %v", err)
			continue
		}

		// Broadcast cho tat ca client trong phong
		h.hub.broadcast <- &BroadcastMsg{
			RoomID:  client.roomID,
			Payload: payload,
		}
	}
}

// ========================================
// WRITE PUMP — gui tin nhan ve client
// ========================================

// writePump gui tin tu send channel -> WebSocket client.
func (h *WSHandler) writePump(client *Client) {
	defer client.conn.Close()

	for msg := range client.send {
		err := client.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("[WS] Write error user %d: %v", client.userID, err)
			break
		}
	}
}
