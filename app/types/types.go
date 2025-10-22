package types

type Message struct {
	Action string `json:"action" validate:"required"`
	RoomID string `json:"roomId,omitempty"`
	Data   string `json:"data,omitempty"`
}

type ColumnEvent struct {
	RoomID    string `json:"room_id"`
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	Name      string `json:"name"`
	Position  int64  `json:"position"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
