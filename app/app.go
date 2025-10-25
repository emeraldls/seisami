package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"seisami/app/internal/actions"
	"seisami/app/internal/repo"
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"

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

const defaultCollabServerAddr = "127.0.0.1:8080"

// App struct
type App struct {
	ctx              context.Context
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
}

// NewApp creates a new App application struct
func NewApp() *App {
	repo := repo.NewRepo()
	addr := os.Getenv("COLLAB_SERVER_ADDR")
	if addr == "" {
		addr = defaultCollabServerAddr
	}
	return &App{
		stopChan:         make(chan bool),
		repository:       repo,
		collabServerAddr: addr,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.action = actions.NewAction(ctx, a.repository)

	go a.handleMutations()
	go startListener()
	go a.handleFnKeyPress()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetBoards(page int64, pageSize int64) ([]query.Board, error) {
	return a.repository.GetAllBoards(page, pageSize)
}

func (a *App) CreateBoard(boardName string) (query.Board, error) {
	return a.repository.CreateBoard(boardName)
}

func (a *App) GetBoardByID(boardId string) (query.Board, error) {
	return a.repository.GetBoard(boardId)
}

func (a *App) UpdateBoard(boardId string, name string) (query.Board, error) {
	return a.repository.UpdateBoard(boardId, name)
}

func (a *App) DeleteBoard(boardId string) error {
	return a.repository.DeleteBoard(boardId)
}

func (a *App) CreateColumn(boardId string, columnName string) (query.Column, error) {
	return a.repository.CreateColumn(boardId, columnName)
}

func (a *App) DeleteColumn(columnId string) error {
	return a.repository.DeleteColumn(columnId)
}

func (a *App) GetColumn(columnId string) (query.Column, error) {
	return a.repository.GetColumn(columnId)
}

func (a *App) ListColumnsByBoard(boardId string) ([]query.Column, error) {
	return a.repository.ListColumnsByBoard(boardId)
}

func (a *App) UpdateColumn(columnId string, name string) (query.Column, error) {
	return a.repository.UpdateColumn(columnId, name)
}

func (a *App) CreateCard(columnId string, title string, description string) (query.Card, error) {
	return a.repository.CreateCard(columnId, title, description)
}

func (a *App) DeleteCard(cardId string) error {
	return a.repository.DeleteCard(cardId)
}

func (a *App) GetCard(CardId string) (query.Card, error) {
	return a.repository.GetCard(CardId)
}

func (a *App) ListCardsByColumn(columnId string) ([]query.Card, error) {
	return a.repository.ListCardsByColumn(columnId)
}

func (a *App) UpdateCard(cardId string, title string, description string) (query.Card, error) {
	return a.repository.UpdateCard(cardId, title, description)
}

func (a *App) UpdateCardColumn(cardId string, columnId string) (query.Card, error) {
	return a.repository.UpdateCardColumn(cardId, columnId)
}

func (a *App) GetTranscriptions(boardId string, page, pageSize int64) ([]query.Transcription, error) {
	return a.repository.GetTranscriptions(boardId, page, pageSize)
}

func (a *App) GetTranscriptionByID(transcriptionId string) (query.Transcription, error) {
	return a.repository.GetTranscriptionByID(transcriptionId)
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

func (a *App) ensureCollabConnectionLocked() error {
	if a.collabConn != nil {
		return nil
	}

	addr := a.collabServerAddr
	if addr == "" {
		addr = defaultCollabServerAddr
	}

	wsURL := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to connect to collaboration server at %s: %w", wsURL.String(), err)
	}

	a.collabConn = conn
	a.collabCloseChan = make(chan bool)

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "collab:connected", map[string]string{"address": addr})
	}

	// Start listening for messages from the server
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

	// block if there's no current roomId

	if a.collabRoomId == "" {
		return nil
	}

	fmt.Println("can send collab command")

	a.collabMu.Lock()
	defer a.collabMu.Unlock()

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

func (a *App) handleFnKeyPress() {
	fnPressed := false

	// Check accessibility permissions once at startup
	accessibilityStatus := checkAccessibilityPermission()
	if accessibilityStatus != 1 {
		fmt.Println("Accessibility permissions not granted - FN key monitoring disabled")
		runtime.EventsEmit(a.ctx, "accessibility:permission_denied", "Accessibility permissions are required for FN key monitoring. Please grant them in System Settings.")
		return
	}

	for {
		if isFnPressed() {
			if !fnPressed {
				fnPressed = true
				fmt.Println("FN key pressed, checking permissions and starting recording")

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

					// Request permission asynchronously but don't block the FN key handler
					go func() {
						if requestMicrophonePermissionSync() {
							fmt.Println("Microphone permission granted via FN key request")
							runtime.EventsEmit(a.ctx, "microphone:permission_granted", "Microphone permission granted. You can now press FN to record.")
						} else {
							fmt.Println("Microphone permission denied via FN key request")
							runtime.EventsEmit(a.ctx, "microphone:permission_denied", "Microphone permission was denied. Please grant permission in System Settings.")
							openMicrophoneSettings()
						}
					}()
				}
			}
		} else {
			if fnPressed {
				fnPressed = false
				fmt.Println("FN key released, stopping recording")
				a.stopRecording()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (a *App) startRecording() {
	PlaySound()
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
	home, _ := os.UserHomeDir()
	appDir := filepath.Join(home, "Music", "Seisami")

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

	var transcription string
	switch settings.TranscriptionMethod {
	case "local":
		transcription, err = a.transcribeWithLocalWhisper(a.recordingPath, settings)
	case "custom":
		transcription, err = a.transcribeWithOpenAI(a.recordingPath, settings)
	case "cloud":
		// TODO: update this when cloud transcription implemented
		fallthrough
	default:
		transcription, err = a.transcribeWithCloud(a.recordingPath)
	}

	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		// TODO: emit error to frontend
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

	// Emit structured response to frontend for display
	runtime.EventsEmit(a.ctx, "structured_response", string(structuredJson))

}

// ExportDataForSync exports all local data for cloud sync
func (a *App) ExportDataForSync() (map[string]interface{}, error) {
	data, err := a.repository.ExportAllData()
	if err != nil {
		return nil, fmt.Errorf("failed to export data: %v", err)
	}

	// Convert to map for JSON marshaling
	return map[string]interface{}{
		"boards":         data.Boards,
		"columns":        data.Columns,
		"cards":          data.Cards,
		"transcriptions": data.Transcriptions,
	}, nil
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
	}

	resp, err := client.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("OpenAI transcription failed: %v", err)
	}

	return resp.Text, nil
}

func (a *App) transcribeWithCloud(filePath string) (string, error) {
	// TODO: implement cloud transcription
	// for now, fall back to local transcription using env API key or current method
	return a.transcribeLocally(filePath)
}
