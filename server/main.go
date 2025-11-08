package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"seisami/server/central"
	"seisami/server/centraldb"
	"seisami/server/client"
	"seisami/server/room_manager"
	"seisami/server/types"
)

// a client needs to create a room
// then when a room has been created, client can join any room using the room id

var roomManager *room_manager.RoomManager
var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

func init() {
	roomManager = room_manager.NewRoomManager()
	central.SetWebSocketHandler(HandleWebSocket)
}

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("unable to load environment variables: %v\n", err)
	}

	shutdownAuth, err := startCentralHTTPServer()
	if err != nil {
		log.Fatalf("unable to start auth HTTP server: %v", err)
	}
	defer shutdownAuth()

	log.Println("Collab Server is running - WebSocket endpoint at ws://0.0.0.0:8080/ws")

	select {}
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	userId := r.Header.Get("X-User-ID")
	boardId := r.Header.Get("X-Board-ID")

	if userId == "" {
		log.Println("WebSocket connection missing X-User-ID header")
		conn.Close()
		return
	}

	if boardId == "" {
		log.Println("WebSocket connection missing X-Board-ID header")
		conn.Close()
		return
	}

	cl := client.NewClient(conn, userId)
	fmt.Println("New client connected:", cl.GetId())

	err = roomManager.JoinOrCreateRoom(boardId, cl)
	if err != nil {
		log.Printf("Failed to join board room: %v", err)
		conn.Close()
		return
	}

	fmt.Printf("Client %s joined board room: %s\n", cl.GetId(), boardId)

	go handleConn(cl, roomManager, boardId)
}

func handleConn(c *client.Client, manager *room_manager.RoomManager, boardId string) {
	defer c.Close()
	defer manager.LeaveRoomById(boardId, c)

	for {
		_, message, err := c.Conn().ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected:", c.GetId())
			return
		}

		var msg types.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			fmt.Println("Invalid JSON:", err)
			response := map[string]string{"error": "invalid JSON"}
			jsonResp, _ := json.Marshal(response)
			c.Send(jsonResp)
			continue
		}

		switch msg.Action {
		case "broadcast":
			broadcastMsg := map[string]string{
				"type": msg.Type,
				"from": c.GetId(),
				"data": msg.Data,
			}
			jsonMsg, _ := json.Marshal(broadcastMsg)
			manager.BroadcastToRoom(boardId, jsonMsg)

		default:
			response := map[string]string{"error": "unknown action"}
			jsonResp, _ := json.Marshal(response)
			c.Send(jsonResp)
		}
	}
}

//go:embed sqlc/schema.sql
var schema string

func startCentralHTTPServer() (func(), error) {
	cfg, err := central.LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	_, err = pool.Exec(ctx, schema)
	if err != nil {
		return nil, fmt.Errorf("unable to create tables: %v", err)
	}

	queries := centraldb.New(pool)
	authService := central.NewAuthService(queries, cfg)
	syncService := central.NewSyncService(pool, queries)
	router := central.NewRouter(authService, syncService)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	go func() {
		log.Printf("Auth HTTP server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("auth HTTP server failed: %v", err)
		}
	}()

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("auth HTTP server shutdown error: %v", err)
		}
		pool.Close()
	}

	return shutdown, nil
}
