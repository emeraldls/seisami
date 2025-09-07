package repo

import (
	"context"
	"fmt"
	"log"
	"os"
	"seisami/sqlc/query"

	_ "embed"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"database/sql"
)

type repo struct {
	queries *query.Queries
	ctx     context.Context
}

func NewRepo() *repo {
	// path := "../../sqlc/schema.sql"
	ctx := context.Background()
	db, err := sql.Open("sqlite3", "seisami.db")
	if err != nil {
		log.Fatalf("unable to setup sqlite: %v\n", err)
	}

	ddl, err := os.ReadFile("sqlc/schema.sql")
	if err != nil {
		log.Fatalf("unable to read schema: %v", err)
	}

	if _, err := db.ExecContext(ctx, string(ddl)); err != nil {
		log.Fatalf("unable to create tables: %v\n", err)
	}

	queries := query.New(db)

	return &repo{
		queries,
		ctx,
	}
}

func (r *repo) GetBoard(boardId string) (query.Board, error) {
	board, err := r.queries.GetBoard(r.ctx, boardId)
	if err != nil {
		return query.Board{}, fmt.Errorf("error occured fetching board: (%s)... %v", boardId, err)
	}

	return board, nil
}

func (r *repo) CreateBoard(name string) (query.Board, error) {
	id := uuid.New().String()
	board, err := r.queries.CreateBoard(r.ctx, query.CreateBoardParams{ID: id, Name: name})
	if err != nil {
		return query.Board{}, fmt.Errorf("error occured creating board: %v", err)
	}

	return board, nil
}

func (r *repo) DeleteBoard(boardId string) error {
	if err := r.queries.DeleteBoard(r.ctx, boardId); err != nil {
		fmt.Println("deleting board...", boardId)
		return fmt.Errorf("error occured deleting board: (%s)...%v", boardId, err)
	}

	return nil
}

func (r *repo) GetAllBoards(page int64, pageSize int64) ([]query.Board, error) {
	boards, err := r.queries.ListBoards(r.ctx, query.ListBoardsParams{Limit: 10, Offset: (page - 1) * pageSize})
	if err != nil {
		return nil, fmt.Errorf("error occured getting all boards: %v", err)
	}

	return boards, nil
}

func (r *repo) UpdateBoard(boardId string, name string) (query.Board, error) {
	return r.queries.UpdateBoard(r.ctx, query.UpdateBoardParams{ID: boardId, Name: name})
}

func (r *repo) CreateColumn(boardId string, columnName string) (query.Column, error) {
	id := uuid.New().String()

	columns, err := r.ListColumnsByBoard(boardId)
	if err != nil {
		return query.Column{}, fmt.Errorf("error getting columns to determine position: %v", err)
	}
	position := len(columns)

	column, err := r.queries.CreateColumn(r.ctx, query.CreateColumnParams{
		ID:       id,
		BoardID:  boardId,
		Title:    columnName,
		Position: int64(position),
	})
	if err != nil {
		return query.Column{}, fmt.Errorf("error creating column: %v", err)
	}
	return column, nil
}

func (r *repo) DeleteColumn(columnId string) error {
	if err := r.queries.DeleteColumn(r.ctx, columnId); err != nil {
		return fmt.Errorf("error deleting column: %v", err)
	}
	return nil
}

func (r *repo) GetColumn(columnId string) (query.Column, error) {
	column, err := r.queries.GetColumn(r.ctx, columnId)
	if err != nil {
		return query.Column{}, fmt.Errorf("error getting column: %v", err)
	}
	return column, nil
}

func (r *repo) ListColumnsByBoard(boardId string) ([]query.Column, error) {
	columns, err := r.queries.ListColumnsByBoard(r.ctx, boardId)
	if err != nil {
		return nil, fmt.Errorf("error listing columns by board: %v", err)
	}
	return columns, nil
}

func (r *repo) UpdateColumn(columnId string, name string) (query.Column, error) {
	column, err := r.queries.UpdateColumn(r.ctx, query.UpdateColumnParams{
		ID:    columnId,
		Title: name,
	})
	if err != nil {
		return query.Column{}, fmt.Errorf("error updating column: %v", err)
	}
	return column, nil
}

func (r *repo) CreateTicket(columnId string, title string, description string, ticketType string) (query.Ticket, error) {
	id := uuid.New().String()
	ticket, err := r.queries.CreateTicket(r.ctx, query.CreateTicketParams{
		ID:       id,
		ColumnID: columnId,
		Title:    title,
		Description: sql.NullString{
			String: description,
			Valid:  true,
		},
		TicketType: ticketType,
	})
	if err != nil {
		return query.Ticket{}, fmt.Errorf("error creating ticket: %v", err)
	}
	return ticket, nil
}

func (r *repo) DeleteTicket(ticketId string) error {
	if err := r.queries.DeleteTicket(r.ctx, ticketId); err != nil {
		return fmt.Errorf("error deleting ticket: %v", err)
	}
	return nil
}

func (r *repo) GetTicket(ticketId string) (query.Ticket, error) {
	ticket, err := r.queries.GetTicket(r.ctx, ticketId)
	if err != nil {
		return query.Ticket{}, fmt.Errorf("error getting ticket: %v", err)
	}
	return ticket, nil
}

func (r *repo) ListTicketsByColumn(columnId string) ([]query.Ticket, error) {
	tickets, err := r.queries.ListTicketsByColumn(r.ctx, columnId)
	if err != nil {
		return nil, fmt.Errorf("error listing tickets by column: %v", err)
	}
	return tickets, nil
}

func (r *repo) UpdateTicket(ticketId string, title string, description string) (query.Ticket, error) {
	ticket, err := r.queries.UpdateTicket(r.ctx, query.UpdateTicketParams{
		ID:    ticketId,
		Title: title,
		Description: sql.NullString{
			String: description,
			Valid:  true,
		},
	})
	if err != nil {
		return query.Ticket{}, fmt.Errorf("error updating ticket: %v", err)
	}
	return ticket, nil
}

func (r *repo) UpdateTicketColumn(ticketId string, columnId string) (query.Ticket, error) {
	ticket, err := r.queries.UpdateTicketColumn(r.ctx, query.UpdateTicketColumnParams{
		ID:       ticketId,
		ColumnID: columnId,
	})
	if err != nil {
		return query.Ticket{}, fmt.Errorf("error updating ticket column: %v", err)
	}
	return ticket, nil
}
