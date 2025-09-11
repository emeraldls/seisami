package repo

import (
	"seisami/sqlc/query"
)

type Repository interface {
	CreateBoard(name string) (query.Board, error)
	DeleteBoard(id string) error
	GetBoard(id string) (query.Board, error)
	// will update this later to include query params
	GetAllBoards(page int64, pageSize int64) ([]query.Board, error)
	UpdateBoard(id string, name string) (query.Board, error)

	CreateColumn(boardId string, columnName string) (query.Column, error)
	DeleteColumn(id string) error
	GetColumn(id string) (query.Column, error)
	ListColumnsByBoard(boardId string) ([]query.Column, error)
	UpdateColumn(id string, name string) (query.Column, error)

	CreateTicket(columnId string, title string, description string, ticketType string) (query.Ticket, error)
	DeleteTicket(id string) error
	GetTicket(id string) (query.Ticket, error)
	ListTicketsByColumn(columnId string) ([]query.Ticket, error)
	UpdateTicket(id string, title string, description string) (query.Ticket, error)
	UpdateTicketColumn(ticketId string, columnId string) (query.Ticket, error)

	AddTransscription(boardId string, transcription string, recordingPath string) (query.Transcription, error)
	GetTranscriptions(boardId string, page, pageSize int64) ([]query.Transcription, error)
	GetTranscriptionByID(transcriptionId string) (query.Transcription, error)
	UpdateTranscriptionIntent(transcriptionId string, intent string) error
	UpdateTranscriptionResponse(transcriptionId string, response string) error
}
