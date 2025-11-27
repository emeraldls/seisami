package cloud

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type SyncWebSocket struct {
	conn            *websocket.Conn
	onSyncUpdate    func(tableName string)
	mu              sync.Mutex
	ctx             context.Context
	cloudApiUrl     string
	sessionToken    string
	isConnected     bool
	shouldReconnect bool
}

type SyncMessage struct {
	Type      string `json:"type"`
	TableName string `json:"table_name,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

func NewSyncWebSocket(ctx context.Context, cloudApiUrl, sessionToken string, onSyncUpdate func(string)) *SyncWebSocket {
	ws := &SyncWebSocket{
		onSyncUpdate:    onSyncUpdate,
		ctx:             ctx,
		cloudApiUrl:     cloudApiUrl,
		sessionToken:    sessionToken,
		shouldReconnect: true,
	}
	return ws
}

func (sw *SyncWebSocket) Connect() error {
	sw.mu.Lock()

	if sw.conn != nil {
		sw.conn.Close()
		sw.conn = nil
	}

	parsedURL, err := url.Parse(sw.cloudApiUrl)
	if err != nil {
		sw.mu.Unlock()
		return fmt.Errorf("invalid cloud URL: %w", err)
	}

	scheme := "wss"
	if parsedURL.Scheme == "http" || os.Getenv("DEV") == "true" {
		scheme = "ws"
	}

	wsURL := url.URL{
		Scheme: scheme,
		Host:   parsedURL.Host,
		Path:   "/sync/ws",
	}
	q := wsURL.Query()
	q.Set("token", sw.sessionToken)
	wsURL.RawQuery = q.Encode()

	log.Printf("Connecting to sync WebSocket: %s", wsURL.String())

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		sw.mu.Unlock()
		return fmt.Errorf("failed to connect to sync WebSocket: %w", err)
	}

	sw.conn = conn
	sw.isConnected = true
	sw.mu.Unlock()

	log.Println("Sync WebSocket connected successfully")

	if sw.conn != nil {
		go sw.readMessages()
	}

	return nil
}

func (sw *SyncWebSocket) Disconnect() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.shouldReconnect = false
	sw.isConnected = false

	if sw.conn != nil {
		sw.conn.Close()
		sw.conn = nil
		log.Println("Sync WebSocket disconnected")
	}
}

func (sw *SyncWebSocket) readMessages() {
	defer func() {
		sw.mu.Lock()
		sw.isConnected = false
		if sw.conn != nil {
			sw.conn.Close()
			sw.conn = nil
		}
		shouldReconnect := sw.shouldReconnect
		sw.mu.Unlock()

		if shouldReconnect {
			log.Println("Connection lost, attempting to reconnect in 5 seconds...")
			time.Sleep(5 * time.Second)
			if err := sw.Connect(); err != nil {
				log.Printf("Reconnection failed: %v", err)
			}
		}
	}()

	for {
		sw.mu.Lock()
		conn := sw.conn
		sw.mu.Unlock()
		if conn == nil {
			log.Println("SyncWebSocket.readMessages: conn is nil, exiting goroutine")
			return
		}

		var msg SyncMessage
		err := conn.ReadJSON(&msg)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Sync WebSocket error: %v", err)
			}
			return
		}

		log.Printf("Sync message received: type=%s, table=%s", msg.Type, msg.TableName)

		if msg.Type == "sync_update" && msg.TableName != "" && sw.onSyncUpdate != nil {
			sw.onSyncUpdate(msg.TableName)
		}
	}
}

func (sw *SyncWebSocket) UpdateToken(token string) {
	sw.mu.Lock()
	sw.sessionToken = token
	sw.mu.Unlock()

	// Reconnect with new token
	sw.Disconnect()
	time.Sleep(500 * time.Millisecond)

	sw.mu.Lock()
	sw.shouldReconnect = true
	sw.mu.Unlock()

	if err := sw.Connect(); err != nil {
		log.Printf("Failed to reconnect with new token: %v", err)
	}
}
