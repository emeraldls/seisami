package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"seisami/app/internal/actions"
	"seisami/app/internal/cloud"
	"seisami/app/internal/repo"
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/internal/sync_engine"
	"seisami/app/types"
	"seisami/app/utils"

	"sort"
	"strings"
	"sync"
	"time"

	"github.com/emeraldls/portaudio"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/gorilla/websocket"
	"github.com/sashabaranov/go-openai"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	_ "embed"
)

/*
TODO: When a user switch a board, kill any active connections
*/

/*
	Whenever im updating card column, card data, or whenever im performing anything on the board, i need to broadcast what the user has done to other clients

	When a new client joins a room that has length > 0, the server sends full board snapshot to the client

	Inviting a client into your own board, which means on their end, a copy of your board will be created & stored.

	TODO: updating of a card name/title isnt implemented.
*/

// App struct
type App struct {
	ctx              context.Context
	loginToken       string
	cloudApiUrl      string
	isRecording      bool
	stopChan         chan bool
	recordingPath    string
	lastBarEmitTime  time.Time
	repository       repo.Repository
	action           *actions.Action
	currentBoardId   string
	collabMu         sync.Mutex
	collabConn       *websocket.Conn
	collabServerAddr string
	collabRoomId     string
	collabCloseChan  chan bool
	cloud            cloud.Cloud
	syncEngine       *sync_engine.SyncEngine
	syncWS           *cloud.SyncWebSocket
}

func dbPath() string {
	dbPath := utils.GetDBPath()

	dbDir := filepath.Dir(dbPath)
	os.MkdirAll(dbDir, 0755)

	return dbPath
}

// NewApp creates a new App application struct
func NewApp() *App {
	db, err := sql.Open("sqlite3", dbPath())
	if err != nil {
		log.Fatalf("unable to setup sqlite: %v\n", err)
	}

	ctx := context.Background()

	if _, err := db.ExecContext(ctx, repo.Schema); err != nil {
		log.Fatalf("unable to create tables: %v\n", err)
	}
	repo := repo.NewRepo(db, ctx)
	addr := getCollabServerAddr()

	return &App{
		stopChan:         make(chan bool),
		repository:       repo,
		collabServerAddr: addr,
		cloudApiUrl:      getCloudApiUrl(),
	}
}

func (a *App) isAuthenticated() bool {
	return strings.TrimSpace(a.loginToken) != ""
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.action = actions.NewAction(ctx, a.repository)

	cloudFuncs := cloud.NewCloudFuncs(a.repository, a.loginToken, a.ctx, a.cloudApiUrl)
	a.cloud = cloudFuncs

	syncEngine := sync_engine.NewSyncEngine(a.repository, cloudFuncs, a.ctx)
	a.syncEngine = syncEngine

	a.syncWS = cloud.NewSyncWebSocket(ctx, a.cloudApiUrl, a.loginToken, func(tableName string) {
		fmt.Printf("Sync update received for table: %s\n", tableName)

		var tableType types.TableName
		switch tableName {
		case "boards":
			tableType = types.BoardTable
		case "columns":
			tableType = types.ColumnTable
		case "cards":
			tableType = types.CardTable
		case "transcriptions":
			tableType = types.TranscriptionTable
		default:
			return
		}

		// Trigger sync for the specific table
		go func() {
			if err := syncEngine.SyncData(tableType, true); err != nil {
				fmt.Printf("Error syncing %s: %v\n", tableName, err)
			} else {
				// Emit event to frontend that data was updated
				runtime.EventsEmit(ctx, "sync:table_updated", map[string]string{
					"table": tableName,
				})
			}
		}()
	})

	go a.handleMutations()
	go a.appVersionCheck()

	a.initPlatformSpecific()
	fmt.Println("is user authenticated: ", a.isAuthenticated())
	if a.isAuthenticated() {
		if err := a.syncWS.Connect(); err != nil {
			fmt.Printf("Failed to connect sync WebSocket: %v\n", err)
		}
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetBoards(page int64, pageSize int64) ([]types.ExportedBoard, error) {
	boards, err := a.repository.GetAllBoards(page, pageSize)

	if err != nil {
		return []types.ExportedBoard{}, err
	}

	var boardResponse = make([]types.ExportedBoard, 0)

	for _, board := range boards {
		boardResponse = append(boardResponse, types.ExportedBoard{
			ID:        board.ID,
			Name:      board.Name,
			CreatedAt: utils.ConvertTimestamptzToLocal(board.CreatedAt),
			UpdatedAt: utils.ConvertTimestamptzToLocal(board.UpdatedAt),
		})
	}

	return boardResponse, nil
}

func (a *App) CreateBoard(boardName string) (types.ExportedBoard, error) {
	createdBoard, err := a.repository.CreateBoard(boardName)
	if err != nil {
		return types.ExportedBoard{}, err
	}

	var board = types.ExportedBoard{
		ID:        createdBoard.ID,
		Name:      createdBoard.Name,
		CreatedAt: utils.ConvertTimestamptzToLocal(createdBoard.CreatedAt),
		UpdatedAt: utils.ConvertTimestamptzToLocal(createdBoard.UpdatedAt),
	}

	return board, nil
}

func (a *App) GetBoardByID(boardId string) (types.ExportedBoard, error) {
	gottenBoard, err := a.repository.GetBoard(boardId)
	if err != nil {
		return types.ExportedBoard{}, err
	}

	var board = types.ExportedBoard{
		ID:        gottenBoard.ID,
		Name:      gottenBoard.Name,
		CreatedAt: utils.ConvertTimestamptzToLocal(gottenBoard.CreatedAt),
		UpdatedAt: utils.ConvertTimestamptzToLocal(gottenBoard.UpdatedAt),
	}

	return board, nil
}

func (a *App) UpdateBoard(boardId string, name string) (types.ExportedBoard, error) {
	updatedBoard, err := a.repository.UpdateBoard(boardId, name)
	if err != nil {
		return types.ExportedBoard{}, err
	}

	var board = types.ExportedBoard{
		ID:        updatedBoard.ID,
		Name:      updatedBoard.Name,
		CreatedAt: utils.ConvertTimestamptzToLocal(updatedBoard.CreatedAt),
		UpdatedAt: utils.ConvertTimestamptzToLocal(updatedBoard.UpdatedAt),
	}

	return board, nil

}

func (a *App) DeleteBoard(boardId string) error {
	return a.repository.DeleteBoard(boardId)
}

func (a *App) CreateColumn(boardId string, columnName string) (types.ExportedColumn, error) {
	column, err := a.repository.CreateColumn(boardId, columnName)
	if err != nil {
		return types.ExportedColumn{}, err
	}

	return types.ExportedColumn{
		ID:        column.ID,
		BoardID:   column.BoardID,
		Name:      column.Name,
		Position:  column.Position,
		CreatedAt: utils.ConvertTimestamptzToLocal(column.CreatedAt),
		UpdatedAt: utils.ConvertTimestamptzToLocal(column.UpdatedAt),
	}, nil
}

func (a *App) DeleteColumn(columnId string) error {
	return a.repository.DeleteColumn(columnId)
}

func (a *App) GetColumn(columnId string) (types.ExportedColumn, error) {
	column, err := a.repository.GetColumn(columnId)
	if err != nil {
		return types.ExportedColumn{}, err
	}

	return types.ExportedColumn{
		ID:        column.ID,
		BoardID:   column.BoardID,
		Name:      column.Name,
		Position:  column.Position,
		CreatedAt: utils.ConvertTimestamptzToLocal(column.CreatedAt),
		UpdatedAt: utils.ConvertTimestamptzToLocal(column.UpdatedAt),
	}, nil
}

func (a *App) ListColumnsByBoard(boardId string) ([]types.ExportedColumn, error) {
	columns, err := a.repository.ListColumnsByBoard(boardId)
	if err != nil {
		return []types.ExportedColumn{}, err
	}

	var columnResponse = make([]types.ExportedColumn, 0)
	for _, column := range columns {
		columnResponse = append(columnResponse, types.ExportedColumn{
			ID:        column.ID,
			BoardID:   column.BoardID,
			Name:      column.Name,
			Position:  column.Position,
			CreatedAt: utils.ConvertTimestamptzToLocal(column.CreatedAt),
			UpdatedAt: utils.ConvertTimestamptzToLocal(column.UpdatedAt),
		})
	}

	return columnResponse, nil
}

func (a *App) UpdateColumn(columnId string, name string) (types.ExportedColumn, error) {
	column, err := a.repository.UpdateColumn(columnId, name)
	if err != nil {
		return types.ExportedColumn{}, err
	}

	return types.ExportedColumn{
		ID:        column.ID,
		BoardID:   column.BoardID,
		Name:      column.Name,
		Position:  column.Position,
		CreatedAt: utils.ConvertTimestamptzToLocal(column.CreatedAt),
		UpdatedAt: utils.ConvertTimestamptzToLocal(column.UpdatedAt),
	}, nil
}

func (a *App) CreateCard(columnId string, title string, description string) (types.ExportedCard, error) {
	card, err := a.repository.CreateCard(columnId, title, description)
	if err != nil {
		return types.ExportedCard{}, err
	}

	var desc string
	if card.Description.Valid {
		desc = card.Description.String
	}

	var attachments string
	if card.Attachments.Valid {
		attachments = card.Attachments.String
	}

	return types.ExportedCard{
		ID:          card.ID,
		ColumnID:    card.ColumnID,
		Title:       card.Title,
		Description: desc,
		Attachments: attachments,
		CreatedAt:   utils.ConvertTimestamptzToLocal(card.CreatedAt),
		UpdatedAt:   utils.ConvertTimestamptzToLocal(card.UpdatedAt),
	}, nil
}

func (a *App) DeleteCard(cardId string) error {
	return a.repository.DeleteCard(cardId)
}

func (a *App) GetCard(CardId string) (types.ExportedCard, error) {
	card, err := a.repository.GetCard(CardId)
	if err != nil {
		return types.ExportedCard{}, err
	}

	var desc string
	if card.Description.Valid {
		desc = card.Description.String
	}

	var attachments string
	if card.Attachments.Valid {
		attachments = card.Attachments.String
	}

	return types.ExportedCard{
		ID:          card.ID,
		ColumnID:    card.ColumnID,
		Title:       card.Title,
		Description: desc,
		Attachments: attachments,
		CreatedAt:   utils.ConvertTimestamptzToLocal(card.CreatedAt),
		UpdatedAt:   utils.ConvertTimestamptzToLocal(card.UpdatedAt),
	}, nil
}

func (a *App) ListCardsByColumn(columnId string) ([]types.ExportedCard, error) {
	cards, err := a.repository.ListCardsByColumn(columnId)
	if err != nil {
		return []types.ExportedCard{}, err
	}

	var cardResponse = make([]types.ExportedCard, 0)
	for _, card := range cards {
		var desc string
		if card.Description.Valid {
			desc = card.Description.String
		}

		var attachments string
		if card.Attachments.Valid {
			attachments = card.Attachments.String
		}

		cardResponse = append(cardResponse, types.ExportedCard{
			ID:          card.ID,
			ColumnID:    card.ColumnID,
			Title:       card.Title,
			Description: desc,
			Attachments: attachments,
			CreatedAt:   utils.ConvertTimestamptzToLocal(card.CreatedAt),
			UpdatedAt:   utils.ConvertTimestamptzToLocal(card.UpdatedAt),
		})
	}

	return cardResponse, nil
}

func (a *App) UpdateCard(cardId string, title string, description string) (types.ExportedCard, error) {
	card, err := a.repository.UpdateCard(cardId, title, description)
	if err != nil {
		return types.ExportedCard{}, err
	}

	var desc string
	if card.Description.Valid {
		desc = card.Description.String
	}

	var attachments string
	if card.Attachments.Valid {
		attachments = card.Attachments.String
	}

	return types.ExportedCard{
		ID:          card.ID,
		ColumnID:    card.ColumnID,
		Title:       card.Title,
		Description: desc,
		Attachments: attachments,
		CreatedAt:   utils.ConvertTimestamptzToLocal(card.CreatedAt),
		UpdatedAt:   utils.ConvertTimestamptzToLocal(card.UpdatedAt),
	}, nil
}

func (a *App) UpdateCardColumn(cardId string, columnId string) (types.ExportedCard, error) {
	card, err := a.repository.UpdateCardColumn(cardId, columnId)
	if err != nil {
		return types.ExportedCard{}, err
	}

	var desc string
	if card.Description.Valid {
		desc = card.Description.String
	}

	var attachments string
	if card.Attachments.Valid {
		attachments = card.Attachments.String
	}

	return types.ExportedCard{
		ID:          card.ID,
		ColumnID:    card.ColumnID,
		Title:       card.Title,
		Description: desc,
		Attachments: attachments,
		CreatedAt:   utils.ConvertTimestamptzToLocal(card.CreatedAt),
		UpdatedAt:   utils.ConvertTimestamptzToLocal(card.UpdatedAt),
	}, nil
}

func (a *App) GetTranscriptions(boardId string, page, pageSize int64) ([]types.ExportedTranscription, error) {
	transcriptions, err := a.repository.GetTranscriptions(boardId, page, pageSize)
	if err != nil {
		return []types.ExportedTranscription{}, err
	}

	var transcriptionResponse = make([]types.ExportedTranscription, 0)
	for _, t := range transcriptions {
		var recordingPath string
		if t.RecordingPath.Valid {
			recordingPath = t.RecordingPath.String
		}

		var intent string
		if t.Intent.Valid {
			intent = t.Intent.String
		}

		var assistantResponse string
		if t.AssistantResponse.Valid {
			assistantResponse = t.AssistantResponse.String
		}

		transcriptionResponse = append(transcriptionResponse, types.ExportedTranscription{
			ID:                t.ID,
			BoardID:           t.BoardID,
			Transcription:     t.Transcription,
			RecordingPath:     recordingPath,
			Intent:            intent,
			AssistantResponse: assistantResponse,
			CreatedAt:         utils.ConvertTimestamptzToLocal(t.CreatedAt),
			UpdatedAt:         utils.ConvertTimestamptzToLocal(t.UpdatedAt),
		})
	}

	return transcriptionResponse, nil
}

func (a *App) GetTranscriptionByID(transcriptionId string) (types.ExportedTranscription, error) {
	t, err := a.repository.GetTranscriptionByID(transcriptionId)
	if err != nil {
		return types.ExportedTranscription{}, err
	}

	var recordingPath string
	if t.RecordingPath.Valid {
		recordingPath = t.RecordingPath.String
	}

	var intent string
	if t.Intent.Valid {
		intent = t.Intent.String
	}

	var assistantResponse string
	if t.AssistantResponse.Valid {
		assistantResponse = t.AssistantResponse.String
	}

	return types.ExportedTranscription{
		ID:                t.ID,
		BoardID:           t.BoardID,
		Transcription:     t.Transcription,
		RecordingPath:     recordingPath,
		Intent:            intent,
		AssistantResponse: assistantResponse,
		CreatedAt:         utils.ConvertTimestamptzToLocal(t.CreatedAt),
		UpdatedAt:         utils.ConvertTimestamptzToLocal(t.UpdatedAt),
	}, nil
}

func (a *App) GetSettings() (query.Setting, error) {
	return a.repository.GetSettings()
}

func (a *App) SaveSettings(transcriptionMethod string, whisperBinaryPath *string, whisperModelPath *string, openaiApiKey *string) (query.Setting, error) {
	return a.repository.CreateOrUpdateSettings(transcriptionMethod, whisperBinaryPath, whisperModelPath, openaiApiKey)
}

func (a *App) OpenFileDialog(title string, filters []runtime.FileFilter) (string, error) {
	options := runtime.OpenDialogOptions{
		Title:   title,
		Filters: filters,
	}
	return runtime.OpenFileDialog(a.ctx, options)
}

func (a *App) bootstrapCloud() error {
	if a.syncEngine == nil {
		return fmt.Errorf("sync engine not initialized")
	}
	return a.syncEngine.BootstrapCloud()
}

func (a *App) SetLoginToken(token string) {
	a.loginToken = token
	a.cloud.UpdateSessionToken(token)

	if a.syncWS != nil {
		a.syncWS.UpdateToken(token)
	}
}

func (a *App) ClearLoginToken() {
	a.loginToken = ""
	a.cloud.UpdateSessionToken("")

	if a.syncWS != nil {
		a.syncWS.Disconnect()
	}

	a.collabMu.Lock()
	a.resetCollabConnectionLocked()
	a.collabRoomId = ""
	a.collabMu.Unlock()
}

func (a *App) ImportNewBoard(boardID string) error {

	return a.syncEngine.ImportNewBoard(boardID)
}

func (a *App) GetLoginToken() string {
	return a.loginToken
}

func (a *App) SetCurrentBoardId(boardId string) {
	a.currentBoardId = boardId
}

func (a *App) GetCurrentBoardId() string {
	return a.currentBoardId
}

func (a *App) triggerSilentSync(tableName types.TableName) {
	if !a.isAuthenticated() || a.syncEngine == nil {
		return
	}

	go func() {
		if err := a.syncEngine.SyncData(tableName, true); err != nil {
			fmt.Printf("silent sync error for %s: %v\n", tableName.String(), err)
		}
	}()
}

func (a *App) appVersionCheck() {
	time.Sleep(20 * time.Second) //just to ensure the app has full started before you run this
	fmt.Println("running app version check")
	localVersion, err := a.repository.GetLocalVersion()
	if err != nil {
		fmt.Println(err)
	}

	cloudResp := a.cloud.FetchAppVersion()
	if cloudResp.Error != "" {
		fmt.Printf("unable to fetch cloud version: %s\n", cloudResp.Error)
		return
	}

	cloud, ok := cloudResp.Data.(types.AppVersion)
	if !ok {
		fmt.Println("invalid app version response type")
		return
	}

	fmt.Printf("local version: %s\n", localVersion)
	fmt.Printf("cloud version: %s\n", cloud.Version)

	if cloud.Version == "" {
		fmt.Println("invalid cloud response: missing version")
		return
	}

	if localVersion == cloud.Version {
		fmt.Println("App is up to date.")
		return
	}

	fmt.Println("New version available:", cloud.Version)

	runtime.EventsEmit(a.ctx, "update:available", map[string]string{
		"version": cloud.Version,
		"notes":   cloud.Notes,
		"url":     cloud.URL,
		"sha256":  cloud.Sha256,
	})

}

func (a *App) InstallUpdate(cloud types.AppVersion) error {
	dest := filepath.Join(os.TempDir(), fmt.Sprintf("seisami-%s.dmg", cloud.Version))
	if err := a.downloadAndVerify(cloud.URL, cloud.Sha256, dest); err != nil {
		return err
	}

	if err := exec.Command("open", dest).Run(); err != nil {
		return fmt.Errorf("failed to open installer: %w", err)
	}

	return a.repository.UpdateLocalVersion(cloud.Version)
}

func (a *App) downloadAndVerify(url, expectedHash, dest string) error {
	emitError := func(message string) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "update:download_error", message)
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		emitError(err.Error())
		return fmt.Errorf("failed to fetch file: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		emitError(resp.Status)
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		emitError(err.Error())
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	total := resp.ContentLength
	if total <= 0 {
		fmt.Println("unknown file size, showing bytes instead of percent")
	}

	if a.ctx != nil {
		totalMB := float64(total) / 1024.0 / 1024.0
		payload := map[string]interface{}{
			"totalBytes": total,
			"totalMB":    totalMB,
		}
		if total <= 0 {
			payload["totalMB"] = nil
		}
		runtime.EventsEmit(a.ctx, "update:download_started", payload)
	}

	hasher := sha256.New()
	mw := io.MultiWriter(out, hasher)

	buf := make([]byte, 32*1024)

	var downloaded int64
	var lastEmit time.Time

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := mw.Write(buf[:n]); wErr != nil {
				emitError(wErr.Error())
				return wErr
			}
			downloaded += int64(n)

			if total > 0 {
				percent := float64(downloaded) / float64(total) * 100
				fmt.Printf("\rDownloading %.2f MB (%.1f%%)", float64(downloaded)/1024/1024, percent)
				if a.ctx != nil {
					now := time.Now()
					if lastEmit.IsZero() || now.Sub(lastEmit) >= 200*time.Millisecond || downloaded == total {
						runtime.EventsEmit(a.ctx, "update:download_progress", map[string]interface{}{
							"downloadedBytes": downloaded,
							"totalBytes":      total,
							"percent":         percent,
							"downloadedMB":    float64(downloaded) / 1024.0 / 1024.0,
							"totalMB":         float64(total) / 1024.0 / 1024.0,
						})
						lastEmit = now
					}
				}
			} else {
				fmt.Printf("\rDownloaded %.2f MB", float64(downloaded)/1024/1024)
				if a.ctx != nil {
					now := time.Now()
					if lastEmit.IsZero() || now.Sub(lastEmit) >= 200*time.Millisecond {
						runtime.EventsEmit(a.ctx, "update:download_progress", map[string]interface{}{
							"downloadedBytes": downloaded,
							"totalBytes":      total,
							"percent":         nil,
							"downloadedMB":    float64(downloaded) / 1024.0 / 1024.0,
							"totalMB":         nil,
						})
						lastEmit = now
					}
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			emitError(err.Error())
			return fmt.Errorf("download failed: %w", err)
		}
	}

	fmt.Println()

	hash := hex.EncodeToString(hasher.Sum(nil))
	if hash != expectedHash {
		emitError("checksum mismatch")
		return fmt.Errorf("sha256 mismatch: expected %s, got %s", expectedHash, hash)
	}

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "update:download_complete", map[string]interface{}{
			"downloadedBytes": downloaded,
			"totalBytes":      total,
		})
	}

	return nil
}

func (a *App) ensureCollabConnectionLocked() error {
	if a.collabConn != nil {
		return nil
	}

	if !a.isAuthenticated() {
		return nil
	}

	addr := getCollabServerAddr()

	dev := os.Getenv("DEV")
	scheme := "wss"
	if dev == "true" {
		scheme = "ws"
	}

	wsURL := url.URL{Scheme: scheme, Host: addr, Path: "/ws"}
	q := wsURL.Query()
	q.Set("token", a.GetLoginToken())
	q.Set("board_id", a.currentBoardId)
	wsURL.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("unable to connect to collaboration server at %s: %w", wsURL.String(), err)
	}

	a.collabConn = conn
	a.collabCloseChan = make(chan bool)

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "collab:connected", map[string]string{"address": addr})
	}

	go a.handleCollabMessages()

	return nil
}

func (a *App) resetCollabConnectionLocked() {
	if a.collabConn != nil {
		_ = a.collabConn.Close()
	}
	if a.collabCloseChan != nil {
		close(a.collabCloseChan)
	}
	a.collabConn = nil
	a.collabCloseChan = nil
}

func (a *App) sendCollabCommand(msg types.Message) error {
	fmt.Println("can send collab command")

	a.collabMu.Lock()
	defer a.collabMu.Unlock()

	if !a.isAuthenticated() {
		return fmt.Errorf("user not authenticated")
	}

	if err := a.ensureCollabConnectionLocked(); err != nil {
		return err
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal collaboration message: %w", err)
	}

	err = a.collabConn.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		a.resetCollabConnectionLocked()
		return fmt.Errorf("unable to send collaboration message: %w", err)
	}

	return nil
}

func (a *App) handleCollabMessages() {
	if a.collabConn == nil {
		return
	}

	for {
		select {
		case <-a.collabCloseChan:
			return
		default:
		}

		_, message, err := a.collabConn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading from WebSocket: %v\n", err)
			a.collabMu.Lock()
			a.resetCollabConnectionLocked()
			a.collabMu.Unlock()
			return
		}

		var response map[string]interface{}
		if err := json.Unmarshal(message, &response); err != nil {
			fmt.Printf("Error unmarshalling WebSocket message: %v\n", err)
			continue
		}

		// Update internal room ID state from server responses
		if status, ok := response["status"].(string); ok {
			switch status {
			case "created", "joined":
				if roomId, ok := response["roomId"].(string); ok {
					a.collabMu.Lock()
					a.collabRoomId = roomId
					a.collabMu.Unlock()
				}
			case "left":
				if roomId, ok := response["roomId"].(string); ok {
					a.collabMu.Lock()
					if a.collabRoomId == roomId {
						a.collabRoomId = ""
					}
					a.collabMu.Unlock()
				}
			}
		}

		// Messages are handled directly by the frontend via WebSocket
		// No need to re-emit as Wails events - the frontend listens directly
	}
}

func (a *App) CreateCollaborationRoom() error {
	err := a.sendCollabCommandAsync(types.Message{Action: "create"})
	if err != nil {
		return err
	}

	return nil
}

func (a *App) sendCollabCommandAsync(msg types.Message) error {
	a.collabMu.Lock()
	defer a.collabMu.Unlock()

	if !a.isAuthenticated() {
		return fmt.Errorf("user not authenticated")
	}

	if err := a.ensureCollabConnectionLocked(); err != nil {
		return err
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal collaboration message: %w", err)
	}

	err = a.collabConn.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		a.resetCollabConnectionLocked()
		return fmt.Errorf("unable to send collaboration message: %w", err)
	}

	return nil
}

func (a *App) JoinCollaborationRoom(roomId string) error {
	if strings.TrimSpace(roomId) == "" {
		return fmt.Errorf("room id is required")
	}

	err := a.sendCollabCommandAsync(types.Message{Action: "join", RoomID: roomId})
	if err != nil {
		return err
	}

	return nil
}

func (a *App) LeaveCollaborationRoom(roomId string) (string, error) {
	if strings.TrimSpace(roomId) == "" {
		return "", fmt.Errorf("room id is required")
	}

	err := a.sendCollabCommandAsync(types.Message{Action: "leave", RoomID: roomId})
	if err != nil {
		return "", err
	}

	// The response will come async and be handled by handleCollabMessages
	return "", nil
}

func (a *App) GetCollaborationRoomId() string {
	a.collabMu.Lock()
	defer a.collabMu.Unlock()
	return a.collabRoomId
}

func (a *App) handleHotkeyPress() {
	hotkeyPressed := false

	// Check accessibility permissions if required by platform
	if requiresAccessibilityPerm {
		accessibilityStatus := checkAccessibilityPermission()
		if accessibilityStatus != 1 {
			fmt.Printf("Accessibility permissions not granted - %s monitoring disabled\n", hotkeyName)
			runtime.EventsEmit(a.ctx, "accessibility:permission_denied",
				fmt.Sprintf("Accessibility permissions are required for %s monitoring. Please grant them in System Settings.", hotkeyName))
			return
		}
	}

	for {
		if isHotkeyPressed() {
			if !hotkeyPressed {
				hotkeyPressed = true
				fmt.Printf("%s pressed, checking permissions and starting recording\n", hotkeyName)

				permissionStatus := checkMicrophonePermission()
				switch permissionStatus {
				case 1:
					// Authorized, proceed with recording
					fmt.Println("Microphone permission authorized, starting recording")
					go a.startRecording()
				case 0:
					// Denied or restricted
					fmt.Println("Microphone permission denied - cannot start recording")
					runtime.EventsEmit(a.ctx, "microphone:permission_denied", "Microphone access is denied. Please grant permission in System Settings to use recording.")
					runtime.EventsEmit(a.ctx, "recording:blocked", "Recording blocked due to missing microphone permissions")
				case -1:
					// Not determined - request permission but don't start recording immediately
					fmt.Println("Microphone permission not determined - requesting permission")
					runtime.EventsEmit(a.ctx, "microphone:requesting_permission", "Microphone permission required. Please grant access when prompted.")

					// Request permission asynchronously but don't block the hotkey handler
					go func() {
						if requestMicrophonePermissionSync() {
							fmt.Println("Microphone permission granted via hotkey request")
							runtime.EventsEmit(a.ctx, "microphone:permission_granted", "Microphone permission granted. You can now press the hotkey to record.")
						} else {
							fmt.Println("Microphone permission denied via hotkey request")
							runtime.EventsEmit(a.ctx, "microphone:permission_denied", "Microphone permission was denied. Please grant permission in System Settings.")
							openMicrophoneSettings()
						}
					}()
				}
			}
		} else {
			if hotkeyPressed {
				hotkeyPressed = false
				fmt.Printf("%s released, stopping recording\n", hotkeyName)
				a.stopRecording()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (a *App) startRecording() {
	playRecordingSound()
	if a.isRecording {
		log.Println("Already recording.")
		return
	}

	a.isRecording = true
	runtime.EventsEmit(a.ctx, "recording:start", true)

	err := portaudio.Initialize()
	if err != nil {
		log.Printf("Unable to initialize portaudio: %v\n", err)
		a.isRecording = false
		return
	}
	defer portaudio.Terminate()

	inputChannels := 1
	outputChannels := 0
	sampleRate := 44100
	framesPerBuffer := 64
	buffer := make([]int16, framesPerBuffer)

	stream, err := portaudio.OpenDefaultStream(inputChannels, outputChannels, float64(sampleRate), len(buffer), &buffer)
	if err != nil {
		log.Printf("Error opening stream: %v\n", err)
		a.isRecording = false
		return
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		log.Printf("Error starting stream: %v\n", err)
		a.isRecording = false
		return
	}
	defer stream.Stop()
	appDir := utils.GetRecordingsDir()

	if err := os.MkdirAll(appDir, 0755); err != nil {
		log.Printf("Error creating recordings folder: %v\n", err)
		a.isRecording = false
		return
	}

	fileName := fmt.Sprintf("recording_%s.wav", time.Now().Format("20060102_150405"))
	fullPath := filepath.Join(appDir, fileName)

	a.recordingPath = fullPath

	outFile, err := os.Create(fullPath)
	if err != nil {
		log.Printf("Error creating file: %v\n", err)
		a.isRecording = false
		return
	}
	defer outFile.Close()

	encoder := wav.NewEncoder(outFile, sampleRate, 16, 1, 1)

	log.Println("Recording...")

	for {
		select {
		case <-a.stopChan:
			log.Println("Stopped recording.")
			a.isRecording = false

			payload := map[string]string{"id": a.recordingPath}
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Unable to serialize recording:stop payload: %v\n", err)
			} else {
				runtime.EventsEmit(a.ctx, "recording:stop", string(payloadBytes))
			}

			if err := encoder.Close(); err != nil {
				log.Printf("Error closing encoder: %v\n", err)
			}
			go a.trancribe()
			return
		default:
			if err := stream.Read(); err != nil {
				log.Printf("Error reading from stream: %v\n", err)
				continue
			}

			now := time.Now()
			if now.Sub(a.lastBarEmitTime) >= 50*time.Millisecond {
				bars := a.getAudioBarsFromBuffer(buffer, 20)
				runtime.EventsEmit(a.ctx, "audio_bars", bars)
				a.lastBarEmitTime = now
			}

			buf := new(audio.IntBuffer)
			buf.Format = &audio.Format{
				NumChannels: 1,
				SampleRate:  sampleRate,
			}
			buf.Data = fromInt16(buffer)
			if err := encoder.Write(buf); err != nil {
				log.Printf("Error writing to WAV file: %v\n", err)
			}
		}
	}

}

func (a *App) stopRecording() {
	if a.isRecording {
		a.stopChan <- true
	}
}

func fromInt16(in []int16) []int {
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = int(v)
	}
	return out
}

func (a *App) getAudioBarsFromBuffer(buffer []int16, barsCount int) []float64 {
	if len(buffer) == 0 {
		return make([]float64, barsCount)
	}

	totalSamples := len(buffer)
	blockSize := totalSamples / barsCount
	if blockSize == 0 {
		blockSize = 1
	}

	bars := make([]float64, barsCount)

	for i := 0; i < barsCount; i++ {
		start := i * blockSize
		end := start + blockSize
		if end > totalSamples {
			end = totalSamples
		}

		sum := 0.0
		for j := start; j < end; j++ {
			sum += math.Abs(float64(buffer[j]))
		}

		sampleCount := end - start
		avg := 0.0
		if sampleCount > 0 {
			avg = sum / float64(sampleCount)
		}
		bars[i] = avg
	}

	// Compute noise floor (e.g., 10th percentile)
	sorted := append([]float64(nil), bars...)
	sort.Float64s(sorted)
	noiseFloor := sorted[int(float64(len(sorted))*0.1)]

	// Normalize with noise floor offset
	maxVal := 0.0
	for _, v := range bars {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal > noiseFloor {
		for i := range bars {
			bars[i] = (bars[i] - noiseFloor) / (maxVal - noiseFloor)
			if bars[i] < 0 {
				bars[i] = 0
			}
		}
	}

	return bars
}

// TODO: I should use another format for emiting data rather than slow json, maybe an array of length2 [id, transcription]
func (a *App) trancribe() {
	file, err := os.Open(a.recordingPath)
	if err != nil {
		fmt.Printf("Error opening WAV file: %v\n", err)
		return
	}
	defer file.Close()

	decoder := wav.NewDecoder(file)
	if decoder == nil {
		fmt.Println("Could not create decoder")
		return
	}

	duration, err := decoder.Duration()
	if err != nil {
		fmt.Printf("Error getting duration: %v\n", err)
		return
	}

	if duration < 1*time.Second {
		fmt.Printf("Recording is too short: %v. Skipping transcription.\n", duration)
		data := map[string]string{
			"id": a.recordingPath,
		}
		dataBytes, _ := json.Marshal(data)
		runtime.EventsEmit(a.ctx, "transcription:short", string(dataBytes))
		return
	}

	settings, err := a.repository.GetSettings()
	if err != nil {
		fmt.Printf("Error getting settings, using default transcription: %v\n", err)
		settings.TranscriptionMethod = "cloud"
	}

	if settings.TranscriptionMethod == "cloud" && !a.isAuthenticated() {
		errData := map[string]string{
			"id":      a.recordingPath,
			"error":   "authentication_required",
			"message": "Please log in to use cloud transcription, or configure your own OpenAI API key in Settings.",
		}
		errBytes, _ := json.Marshal(errData)
		runtime.EventsEmit(a.ctx, "transcription:error", string(errBytes))
		fmt.Println("Cloud transcription requires authentication")
		return
	}

	if settings.TranscriptionMethod == "custom" && (!settings.OpenaiApiKey.Valid || settings.OpenaiApiKey.String == "") {
		errData := map[string]string{
			"id":      a.recordingPath,
			"error":   "api_key_required",
			"message": "Please configure your OpenAI API key in Settings to use custom transcription.",
		}
		errBytes, _ := json.Marshal(errData)
		runtime.EventsEmit(a.ctx, "transcription:error", string(errBytes))
		fmt.Println("Custom transcription requires OpenAI API key")
		return
	}

	var transcription string
	switch settings.TranscriptionMethod {
	case "local":
		transcription, err = a.transcribeWithLocalWhisper(a.recordingPath, settings)
	case "custom":
		transcription, err = a.transcribeWithOpenAI(a.recordingPath, settings)
	case "cloud":
		cloudData := map[string]string{
			"recording_path": a.recordingPath,
			"board_id":       a.currentBoardId,
		}
		cloudBytes, _ := json.Marshal(cloudData)
		runtime.EventsEmit(a.ctx, "transcription:use_cloud", string(cloudBytes))
		return
	default:
		transcription, err = a.transcribeWithCloud(a.recordingPath)
	}

	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		errData := map[string]string{
			"id":      a.recordingPath,
			"error":   "transcription_failed",
			"message": fmt.Sprintf("Transcription failed: %v", err),
		}
		errBytes, _ := json.Marshal(errData)
		runtime.EventsEmit(a.ctx, "transcription:error", string(errBytes))
		return
	}

	// TODO: adjust db for transcription error
	if strings.Contains(transcription, "[BLANK AUDIO") {

	}

	transcriptionRecord, err := a.repository.AddTransscription(a.currentBoardId, transcription, a.recordingPath)
	if err != nil {
		fmt.Println("unable to create transcription: ", err)
		return
	}

	data := map[string]string{
		"id":            a.recordingPath,
		"transcription": transcription,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Unable to serialize transcription map: %v\n", err)
		return
	}

	runtime.EventsEmit(a.ctx, "transcription", string(dataBytes))

	result, err := a.action.ProcessTranscription(transcription, a.currentBoardId)
	if err != nil {
		fmt.Println("unable to process transcription:", err)
		return
	}

	if result.Intent != "" {
		err = a.repository.UpdateTranscriptionIntent(transcriptionRecord.ID, result.Intent)
		if err != nil {
			fmt.Println("unable to update transcription intent:", err)
		}
	}

	if result.Result != "" {
		err = a.repository.UpdateTranscriptionResponse(transcriptionRecord.ID, result.Result)
		if err != nil {
			fmt.Println("unable to update transcription response:", err)
		}
	}

	structuredJson, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		fmt.Println("unable to marshal structured response:", err)
		return
	}

	fmt.Println("Structured Response:", string(structuredJson))

	runtime.EventsEmit(a.ctx, "structured_response", string(structuredJson))
}

func (a *App) ReprocessTranscription(transcriptionId string, transcriptionText string, boardId string) error {
	if strings.TrimSpace(transcriptionText) == "" {
		return fmt.Errorf("transcription text is empty")
	}

	if strings.TrimSpace(boardId) == "" {
		return fmt.Errorf("board ID is required")
	}

	runtime.EventsEmit(a.ctx, "ai:processing_start", map[string]string{
		"transcriptionId": transcriptionId,
	})

	result, err := a.action.ProcessTranscription(transcriptionText, boardId)
	if err != nil {
		errMsg := fmt.Sprintf("unable to reprocess transcription: %v", err)
		runtime.EventsEmit(a.ctx, "ai:error", map[string]string{
			"error":           errMsg,
			"transcriptionId": transcriptionId,
		})
		return fmt.Errorf("%s", errMsg)
	}

	if result.Intent != "" {
		err = a.repository.UpdateTranscriptionIntent(transcriptionId, result.Intent)
		if err != nil {
			fmt.Printf("unable to update transcription intent: %v\n", err)
		}
	}

	if result.Result != "" {
		err = a.repository.UpdateTranscriptionResponse(transcriptionId, result.Result)
		if err != nil {
			fmt.Printf("unable to update transcription response: %v\n", err)
		}
	}

	structuredJson, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		fmt.Printf("unable to marshal structured response: %v\n", err)
		return fmt.Errorf("failed to serialize response")
	}

	fmt.Println("Reprocessed Structured Response:", string(structuredJson))

	runtime.EventsEmit(a.ctx, "ai:processing_complete", map[string]interface{}{
		"transcriptionId": transcriptionId,
		"intent":          result.Intent,
		"result":          result.Result,
	})

	runtime.EventsEmit(a.ctx, "structured_response", string(structuredJson))

	return nil
}

func (a *App) transcribeLocally(filePath string) (string, error) {
	cmd := exec.Command("/Users/lawrenceishim/Desktop/C/whisper.cpp/build/bin/whisper-cli", "-m", "/Users/lawrenceishim/Desktop/C/whisper.cpp/models/ggml-base.en.bin",
		"-f", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running whisper-cli: %v\n%s", err, string(output))
	}

	lines := strings.Split(string(output), "\n")
	var transcriptionLines []string
	for _, line := range lines {
		if strings.Contains(line, "-->") {
			parts := strings.Split(line, "]")
			if len(parts) > 1 {
				transcriptionLines = append(transcriptionLines, strings.TrimSpace(parts[1]))
			}
		}
	}

	if len(transcriptionLines) > 0 {
		return strings.Join(transcriptionLines, " "), nil
	}

	return strings.TrimSpace(string(output)), nil
}

func (a *App) transcribeWithLocalWhisper(filePath string, settings query.Setting) (string, error) {
	binaryPath := ""
	modelPath := ""

	if settings.WhisperBinaryPath.Valid && settings.WhisperBinaryPath.String != "" {
		binaryPath = settings.WhisperBinaryPath.String
	}
	if settings.WhisperModelPath.Valid && settings.WhisperModelPath.String != "" {
		modelPath = settings.WhisperModelPath.String
	}

	cmd := exec.Command(binaryPath, "-m", modelPath, "-f", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running whisper-cli: %v\n%s", err, string(output))
	}

	lines := strings.Split(string(output), "\n")
	var transcriptionLines []string
	for _, line := range lines {
		if strings.Contains(line, "-->") {
			parts := strings.Split(line, "]")
			if len(parts) > 1 {
				transcriptionLines = append(transcriptionLines, strings.TrimSpace(parts[1]))
			}
		}
	}

	if len(transcriptionLines) > 0 {
		return strings.Join(transcriptionLines, " "), nil
	}

	return strings.TrimSpace(string(output)), nil
}

func (a *App) transcribeWithOpenAI(filePath string, settings query.Setting) (string, error) {
	if !settings.OpenaiApiKey.Valid || settings.OpenaiApiKey.String == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	client := openai.NewClient(settings.OpenaiApiKey.String)

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filePath,
		Language: "en",
	}

	resp, err := client.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("OpenAI transcription failed: %v", err)
	}

	return resp.Text, nil
}

func (a *App) transcribeWithCloud(filePath string) (string, error) {
	// Cloud transcription is handled directly by the frontend
	// The frontend will send audio directly to the cloud API with SSE
	return "", fmt.Errorf("cloud transcription is handled by the frontend - this should not be called")
}

func (a *App) GetCollabServerAddress() string {
	return getCollabServerAddr()
}

func (a *App) GetCloudAPIURL() string {
	return getCloudApiUrl()
}

func (a *App) GetWebURL() string {
	return getWebUrl()
}

type AudioResponse struct {
	Data []byte `json:"data"`
}

func (a *App) ReadAudioFile(filePath string) (*AudioResponse, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	fmt.Println("Length of audio data: ", len(data))

	return &AudioResponse{Data: data}, nil

}
