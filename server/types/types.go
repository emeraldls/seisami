package types

type Message struct {
	Action string `json:"action" validate:"required"`
	RoomID string `json:"roomId,omitempty"`
	Data   string `json:"data,omitempty"`
	Type   string `json:"type"`
}

type SyncBoard struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type SyncColumn struct {
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	Name      string `json:"name"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type SyncCard struct {
	ID          string `json:"id"`
	ColumnID    string `json:"column_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Attachments string `json:"attachments,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type SyncTranscription struct {
	ID                string `json:"id"`
	BoardID           string `json:"board_id"`
	Transcription     string `json:"transcription"`
	RecordingPath     string `json:"recording_path,omitempty"`
	Intent            string `json:"intent,omitempty"`
	AssistantResponse string `json:"assistant_response,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type SyncPayload struct {
	Boards         []SyncBoard         `json:"boards"`
	Columns        []SyncColumn        `json:"columns"`
	Cards          []SyncCard          `json:"cards"`
	Transcriptions []SyncTranscription `json:"transcriptions"`
}

type SyncResponse struct {
	Status           string `json:"status"` // "pending", "in_progress", "completed", "error"
	TotalItems       int    `json:"total_items"`
	ProcessedItems   int    `json:"processed_items"`
	FailedItems      int    `json:"failed_items"`
	Message          string `json:"message,omitempty"`
	Error            string `json:"error,omitempty"`
	DuplicateBoards  int    `json:"duplicate_boards,omitempty"`
	DuplicateColumns int    `json:"duplicate_columns,omitempty"`
	DuplicateCards   int    `json:"duplicate_cards,omitempty"`
}

type SyncStatePayload struct {
	TableName      string `json:"table_name"`
	LastSyncedAt   int64  `json:"last_synced_at"`
	LastSyncedOpID string `json:"last_synced_operation_id"`
}

type BoardRole int

const (
	BoardOwnerRole BoardRole = iota + 1
	BoardMemberRole
)

func (b BoardRole) String() string {
	return [...]string{"owner", "member"}[b-1]
}

type ExportedBoard struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ExportedColumn struct {
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	Name      string `json:"name"`
	Position  int64  `json:"position"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ExportedCard struct {
	ID          string `json:"id"`
	ColumnID    string `json:"column_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Attachments string `json:"attachments,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ExportedTranscription struct {
	ID                string `json:"id"`
	BoardID           string `json:"board_id"`
	Transcription     string `json:"transcription"`
	RecordingPath     string `json:"recording_path,omitempty"`
	Intent            string `json:"intent,omitempty"`
	AssistantResponse string `json:"assistant_response,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type ExportedData struct {
	Board          ExportedBoard           `json:"board"`
	Columns        []ExportedColumn        `json:"columns"`
	Cards          []ExportedCard          `json:"cards"`
	Transcriptions []ExportedTranscription `json:"transcriptions"`
}

type ExportedAllData struct {
	Boards         []ExportedBoard         `json:"boards"`
	Columns        []ExportedColumn        `json:"columns"`
	Cards          []ExportedCard          `json:"cards"`
	Transcriptions []ExportedTranscription `json:"transcriptions"`
}

type BoardMember struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
	Email    string `json:"email"`
}

type BoardMetadata struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	ColumnsCount        int    `json:"columns_count"`
	CardsCount          int    `json:"cards_count"`
	TranscriptionsCount int    `json:"transcriptions_count"`
}
type AppVersion struct {
	Version string `json:"version"`
	Notes   string `json:"notes"`
	URL     string `json:"url"`
	Sha256  string `json:"sha256"`
}
