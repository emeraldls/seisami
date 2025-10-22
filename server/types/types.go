package types

type Message struct {
	Action string `json:"action" validate:"required"`
	RoomID string `json:"roomId,omitempty"`
	Data   string `json:"data,omitempty"`
}
