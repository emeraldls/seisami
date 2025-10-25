package client

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type State int

const (
	Idle State = iota + 1
	InRoom
)

func (s State) String() string {
	return [...]string{"idle", "in_room"}[s-1]
}

type Client struct {
	state State
	conn  *websocket.Conn
	id    string
	mu    sync.Mutex
}

func NewClient(conn *websocket.Conn) *Client {
	id := uuid.New().String()
	return &Client{
		conn:  conn,
		state: Idle,
		id:    id,
	}
}

func (c *Client) GetId() string {
	return c.id
}

func (c *Client) Conn() *websocket.Conn {
	return c.conn
}

func (c *Client) GetState() State {
	return c.state
}

func (c *Client) UpdateState(state State) {
	c.state = state
}

func (c *Client) Send(msg []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.Close()
}
