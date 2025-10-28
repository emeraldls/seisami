package types

import "fmt"

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

type CardEvent struct {
	Column ColumnEvent `json:"column"`
	Card   struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ColumnID    string `json:"column_id"`
		Index       int    `json:"index"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}
}

type CardColumnEvent struct {
	CardID    string `json:"card_id"`
	RoomID    string `json:"room_id"`
	CardIndex int    `json:"card_index"`

	OldColumn struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Position int    `json:"position"`
	} `json:"old_column"`
	NewColumn struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Position int    `json:"position"`
	} `json:"new_column"`
}

type Operation int

const (
	// Basic Row Operation
	InsertOperation Operation = iota + 1
	UpdateOperation
	DeleteOperation

	// Higher Operations
	UpdateCardColumn
)

func (o Operation) String() string {
	return [...]string{"insert", "update", "delete", "update-card-column"}[o-1]
}

type TableName int

const (
	BoardTable TableName = iota + 1
	ColumnTable
	CardTable
	TranscriptionTable
)

func (t TableName) String() string {
	return [...]string{"boards", "columns", "cards", "transcriptions"}[t-1]
}

func TableNameFromString(s string) (TableName, error) {
	switch s {
	case "boards":
		return BoardTable, nil
	case "columns":
		return ColumnTable, nil
	case "cards":
		return CardTable, nil
	case "transcriptions":
		return TranscriptionTable, nil
	default:
		return 0, fmt.Errorf("unknown table name: %s", s)
	}
}

type OperationSync struct {
	ID            string `json:"id"`
	TableName     string `json:"table_name"`
	RecordID      string `json:"record_id"`
	OperationType string `json:"operation_type"`
	DeviceID      string `json:"device_id"`
	PayloadData   string `json:"payload"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type SyncStatePayload struct {
	TableName      string `json:"table_name"`
	LastSyncedAt   int64  `json:"last_synced_at"`
	LastSyncedOpID string `json:"last_synced_operation_id"`
}
