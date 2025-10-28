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

	cl := client.NewClient(conn)
	fmt.Println("New client connected:", cl.GetId())

	go handleConn(cl, roomManager)
}

func handleConn(c *client.Client, manager *room_manager.RoomManager) {
	defer c.Close()

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
		case "join":
			err := manager.JoinRoomById(msg.RoomID, c)
			if err != nil {
				response := map[string]string{"error": fmt.Sprintf("failed to join room: %v", err)}
				jsonResp, _ := json.Marshal(response)
				c.Send(jsonResp)
			} else {
				response := map[string]string{"status": "joined", "roomId": msg.RoomID}
				jsonResp, _ := json.Marshal(response)
				c.Send(jsonResp)
			}
		case "create":
			room := manager.CreateRoom()
			response := map[string]string{"status": "created", "roomId": room.GetRoomId()}
			jsonResp, _ := json.Marshal(response)
			c.Send(jsonResp)

		case "leave":
			manager.LeaveRoomById(msg.RoomID, c)
			response := map[string]string{"status": "left", "roomId": msg.RoomID}
			jsonResp, _ := json.Marshal(response)
			c.Send(jsonResp)

		case "broadcast":
			broadcastMsg := map[string]string{
				"type": "message",
				"from": c.GetId(),
				"data": msg.Data,
			}
			jsonMsg, _ := json.Marshal(broadcastMsg)
			manager.BroadcastToRoom(msg.RoomID, jsonMsg)

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
	syncService := central.NewSyncService(pool)
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
