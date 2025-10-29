package types

type Message struct {
	Action string `json:"action" validate:"required"`
	RoomID string `json:"roomId,omitempty"`
	Data   string `json:"data,omitempty"`
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

type SyncSettings struct {
	ID                  string `json:"id"`
	TranscriptionMethod string `json:"transcription_method"`
	WhisperBinaryPath   string `json:"whisper_binary_path,omitempty"`
	WhisperModelPath    string `json:"whisper_model_path,omitempty"`
	OpenaiAPIKey        string `json:"openai_api_key,omitempty"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type SyncPayload struct {
	Boards         []SyncBoard         `json:"boards"`
	Columns        []SyncColumn        `json:"columns"`
	Cards          []SyncCard          `json:"cards"`
	Transcriptions []SyncTranscription `json:"transcriptions"`
	Settings       *SyncSettings       `json:"settings,omitempty"`
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
