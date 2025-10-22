package main

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

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

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("unable to load environment variables: %v\n", err)
	}

	shutdownAuth, err := startCentralHTTPServer()
	if err != nil {
		log.Fatalf("unable to start auth HTTP server: %v", err)
	}
	defer shutdownAuth()

	list, err := net.Listen("tcp", "0.0.0.0:2121")
	if err != nil {
		log.Fatalf("unable to setup listener: %v\n", err)
	}

	manager := room_manager.NewRoomManager()

	log.Println("Collab Server is running at 2121")

	for {
		conn, err := list.Accept()
		if err != nil {
			log.Fatalf("unable to accept connections: %v\n", err)
		}

		cl := client.NewClient(conn)

		go handleConn(cl, manager)
	}

}

func handleConn(c *client.Client, manager *room_manager.RoomManager) {
	defer c.Close()
	fmt.Println("New client connected:", c.GetId())

	reader := bufio.NewReader(c.Conn())
	for {
		// Read until newline (\n) so clients can send multiple messages
		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println("Client disconnected:", c.GetId())
			return
		}

		var msg types.Message
		if err := json.Unmarshal(line, &msg); err != nil {
			fmt.Println("Invalid JSON:", err)
			continue
		}

		switch msg.Action {
		case "join":
			err := manager.JoinRoomById(msg.RoomID, c)
			if err != nil {
				c.Send([]byte(fmt.Sprintf("error: %v\n", err)))
			} else {
				c.Send([]byte("joined " + msg.RoomID + "\n"))
			}
		case "create":
			room := manager.CreateRoom()
			c.Send([]byte("created " + room.GetRoomId() + "\n"))

		case "leave":
			manager.LeaveRoomById(msg.RoomID, c)
			c.Send([]byte("left " + msg.RoomID + "\n"))

			// TODO: bug - a user that's not in a room can broadcast if he has the room id
		case "broadcast":
			manager.BroadcastToRoom(msg.RoomID, []byte(fmt.Sprintf("[%s]: %s\n", c.GetId(), msg.Data)))

		default:
			c.Send([]byte("unknown action\n"))
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
	service := central.NewAuthService(queries, cfg)
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: central.NewRouter(service),
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
