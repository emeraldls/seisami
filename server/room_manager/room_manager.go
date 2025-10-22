package room_manager

import (
	"errors"
	"sync"

	"seisami/server/client"
	"seisami/server/room"
)

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*room.Room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*room.Room),
	}
}

func (m *RoomManager) CreateRoom() *room.Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	room := room.NewRoom()
	m.rooms[room.GetRoomId()] = room
	return room
}

func (m *RoomManager) GetRoom(roomId string) (*room.Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, exists := m.rooms[roomId]
	if !exists {
		return nil, errors.New("room not found")
	}
	return room, nil
}

func (m *RoomManager) DeleteRoom(roomId string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.rooms, roomId)
}

func (m *RoomManager) JoinRoomById(roomId string, c *client.Client) error {
	room, err := m.GetRoom(roomId)
	if err != nil {
		return err
	}

	err = room.JoinRoom(c)
	if err == nil {
		c.UpdateState(client.InRoom)
	}
	return err
}

func (m *RoomManager) LeaveRoomById(roomId string, c *client.Client) error {
	room, err := m.GetRoom(roomId)
	if err != nil {
		return err
	}

	err = room.LeaveRoom(c)
	if err == nil {
		c.UpdateState(client.Idle)
	}
	return err
}

func (m *RoomManager) BroadcastToRoom(roomId string, msg []byte) error {
	room, err := m.GetRoom(roomId)
	if err != nil {
		return err
	}

	room.Broadcast(msg)
	return nil
}
