package client

import (
	"net"
	"sync"

	"github.com/google/uuid"
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
	conn  net.Conn
	id    string
	mu    sync.Mutex
}

func NewClient(conn net.Conn) *Client {
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

func (c *Client) Conn() net.Conn {
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

	_, err := c.conn.Write(msg)
	return err
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.Close()
}
