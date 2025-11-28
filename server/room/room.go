package room

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"seisami/server/client"

	"github.com/google/uuid"
)

type Room struct {
	id      string
	clients []*client.Client
	mu      sync.RWMutex
}

func NewRoom() *Room {
	return &Room{
		id:      uuid.NewString(),
		clients: make([]*client.Client, 0),
	}
}

func NewRoomWithID(id string) *Room {
	return &Room{
		id:      id,
		clients: make([]*client.Client, 0),
	}
}

func (r *Room) GetRoomId() string {
	return r.id
}

func (r *Room) JoinRoom(client *client.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if slices.Contains(r.clients, client) {
		return errors.New("user already in room")
	}

	r.clients = append(r.clients, client)
	return nil
}

func (r *Room) LeaveRoom(client *client.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := slices.Index(r.clients, client)
	if idx == -1 {
		return errors.New("user not in room")
	}

	r.clients = append(r.clients[:idx], r.clients[idx+1:]...)
	return nil
}

// Broadcast sends a message to all clients in the room.
// You can adapt this based on your WebSocket or message system.
func (r *Room) Broadcast(message []byte, senderId string) {

	fmt.Println("----- The message is -------: ", string(message))

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, client := range r.clients {
		if client.GetId() != senderId {
			client.Send(message)
		}
	}
}

func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

func (r *Room) GetClients() []*client.Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clientsCopy := make([]*client.Client, len(r.clients))
	copy(clientsCopy, r.clients)
	return clientsCopy
}

func (r *Room) CloseRoom() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, client := range r.clients {
		client.Close()
	}

	r.clients = nil
}
