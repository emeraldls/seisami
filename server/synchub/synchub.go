package synchub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type SyncClient struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
}

type SyncHub struct {
	clients    map[string]*SyncClient
	register   chan *SyncClient
	unregister chan *SyncClient
	broadcast  chan SyncMessage
	mu         sync.RWMutex
}

type SyncMessage struct {
	Type      string      `json:"type"`
	TableName string      `json:"table_name,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func NewSyncHub() *SyncHub {
	return &SyncHub{
		clients:    make(map[string]*SyncClient),
		register:   make(chan *SyncClient),
		unregister: make(chan *SyncClient),
		broadcast:  make(chan SyncMessage, 256),
	}
}

func (h *SyncHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("Sync client registered: %s", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Sync client unregistered: %s", client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			client, ok := h.clients[message.UserID]
			h.mu.RUnlock()

			if ok {
				select {
				case client.Send <- mustMarshal(message):
				default:
					h.mu.Lock()
					delete(h.clients, client.UserID)
					close(client.Send)
					h.mu.Unlock()
				}
			}
		}
	}
}

func (h *SyncHub) NotifyUserSync(userID, tableName string) {
	log.Printf("NotifyUserSync called: userID=%s, tableName=%s", userID, tableName)

	h.mu.RLock()
	_, exists := h.clients[userID]
	h.mu.RUnlock()

	if !exists {
		log.Printf("No WebSocket client found for user: %s", userID)
		return
	}

	log.Printf("Broadcasting sync update to user %s for table %s", userID, tableName)

	h.broadcast <- SyncMessage{
		Type:      "sync_update",
		TableName: tableName,
		UserID:    userID,
	}
}

func (h *SyncHub) Register(client *SyncClient) {
	h.register <- client
}

func (h *SyncHub) Unregister(client *SyncClient) {
	h.unregister <- client
}

func (c *SyncClient) ReadPump(hub *SyncHub) {
	defer func() {
		hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("sync websocket error: %v", err)
			}
			break
		}
	}
}

func (c *SyncClient) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Printf("Sending sync message to user %s: %s", c.UserID, string(message))

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing message to user %s: %v", c.UserID, err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return []byte("{}")
	}
	return data
}

var globalHub *SyncHub

func Init() *SyncHub {
	globalHub = NewSyncHub()
	go globalHub.Run()
	return globalHub
}

func Get() *SyncHub {
	return globalHub
}
