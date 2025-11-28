package main

import (
	"context"
	"encoding/json"
	"fmt"
	"seisami/app/types"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) handleMutations() {

	runtime.EventsOn(a.ctx, "auth:set_token", func(optionalData ...any) {
		if len(optionalData) > 0 {
			if token, ok := optionalData[0].(string); ok && token != "" {
				fmt.Printf("Received auth token, setting up cloud authentication\n")
				if a.cloud != nil {
					a.cloud.UpdateSessionToken(token)
				}
			} else {
				fmt.Println("Received event 'auth:set_token' with invalid token")
			}
		} else {
			fmt.Println("Received event 'auth:set_token' with no data")
		}
	})

	runtime.EventsOn(a.ctx, "auth:set_current_board", func(optionalData ...any) {
		if len(optionalData) > 0 {
			if boardId, ok := optionalData[0].(string); ok && boardId != "" {
				fmt.Printf("Received current board ID, setting up collaboration\n")
				a.SetCurrentBoardId(boardId)
			} else {
				fmt.Println("Received event 'auth:set_current_board' with invalid board ID")
			}
		} else {
			fmt.Println("Received event 'auth:set_current_board' with no data")
		}
	})

	runtime.EventsOn(a.ctx, "board:id", func(optionalData ...any) {

		if len(optionalData) > 0 {
			if boardId, ok := optionalData[0].(string); ok {
				a.currentBoardId = boardId
				fmt.Printf("Received event 'board:id' with data: %s\n", boardId)
			}
		} else {
			fmt.Println("Received event 'board:id' with no data")
		}
	})

	runtime.EventsOn(a.ctx, "board:create", func(optionalData ...any) {
		if len(optionalData) > 0 {
			data, ok := optionalData[0].(string)
			if !ok {
				fmt.Println("board:create event received invalid data type")
				return
			}

			go func(ctx context.Context, payload string) {
				select {
				case <-ctx.Done():
					return
				default:
					var boardData map[string]interface{}
					if err := json.Unmarshal([]byte(payload), &boardData); err != nil {
						fmt.Printf("unable to unmarshal board data: %v\n", err)
						return
					}

					boardID, ok := boardData["id"].(string)
					if !ok || boardID == "" {
						fmt.Println("board:create event missing board id")
						return
					}

					_, err := a.repository.CreateOperation(types.BoardTable, boardID, payload, types.InsertOperation)
					if err != nil {
						fmt.Printf("unable to create board operation: %v\n", err)
					}
				}
			}(a.ctx, data)
		}
	})

	runtime.EventsOn(a.ctx, "board:data", func(optionalData ...any) {
		if len(optionalData) > 0 {
			data, ok := optionalData[0].(string)
			if !ok {
				fmt.Println("board:data event received invalid data type")
				return
			}

			go func(ctx context.Context, payload string) {
				select {
				case <-ctx.Done():
					return
				default:
					var boardData map[string]interface{}
					if err := json.Unmarshal([]byte(payload), &boardData); err != nil {
						fmt.Printf("unable to unmarshal board data: %v\n", err)
						return
					}

					boardID, ok := boardData["id"].(string)
					if !ok || boardID == "" {
						fmt.Println("board:data event missing board id")
						return
					}

					_, err := a.repository.CreateOperation(types.BoardTable, boardID, payload, types.UpdateOperation)
					if err != nil {
						fmt.Printf("unable to update board operation: %v\n", err)
						return
					}

				}
			}(a.ctx, data)
		}
	})

	// column name updatedd....
	runtime.EventsOn(a.ctx, "column:data", func(optionalData ...any) {
		if len(optionalData) > 0 {
			var columnData types.ColumnEvent
			data, ok := optionalData[0].(string)

			if ok {
				if err := json.Unmarshal([]byte(data), &columnData); err != nil {
					// TODO: emit error
					fmt.Printf("unable to unmarshal json: %v\n", err)
					return
				}
			} else {
				fmt.Printf("couldnt parse data structure")
				return
			}

			columnBytes, err := json.Marshal(columnData)
			if err != nil {
				fmt.Println("unable to serialize column data: ", err)
				return
			}

			go func(ctx context.Context, column types.ColumnEvent) {
				select {
				case <-ctx.Done():
					return
				default:

					_, err := a.repository.CreateOperation(types.ColumnTable, columnData.ID, string(columnBytes), types.UpdateOperation)
					if err != nil {
						fmt.Println(err)
						return
					}

				}
			}(a.ctx, columnData)

		}
	})

	runtime.EventsOn(a.ctx, "column:create", func(optionalData ...any) {
		if len(optionalData) > 0 {
			var columnData types.ColumnEvent
			data, ok := optionalData[0].(string)
			if ok {
				if err := json.Unmarshal([]byte(data), &columnData); err != nil {
					fmt.Printf("unable to unmarshal json: %v\n", err)
					return
				}
			} else {
				fmt.Printf("couldnt parse data structure")
				return
			}

			go func(ctx context.Context, payload string, column types.ColumnEvent) {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := a.repository.CreateOperation(types.ColumnTable, column.ID, payload, types.InsertOperation)
					if err != nil {
						fmt.Println(err)
					}
				}
			}(a.ctx, data, columnData)

		}
	})

	runtime.EventsOn(a.ctx, "column:delete", func(optionalData ...any) {
		if len(optionalData) > 0 {
			var columnData types.ColumnDeleteEvent
			data, ok := optionalData[0].(string)
			if ok {
				if err := json.Unmarshal([]byte(data), &columnData); err != nil {
					fmt.Printf("unable to unmarshal json: %v\n", err)
					return
				}
			} else {
				fmt.Printf("couldnt parse data structure")
				return
			}

			go func(ctx context.Context, payload string, column types.ColumnDeleteEvent) {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := a.repository.CreateOperation(types.ColumnTable, column.ID, payload, types.DeleteOperation)
					if err != nil {
						fmt.Println(err)
					}
				}
			}(a.ctx, data, columnData)

		}
	})

	runtime.EventsOn(a.ctx, "card:data", func(optionalData ...interface{}) {
		if len(optionalData) > 0 {
			var cardData types.CardEvent
			data, ok := optionalData[0].(string)

			if ok {
				if err := json.Unmarshal([]byte(data), &cardData); err != nil {
					// TODO: emit error
					fmt.Printf("unable to unmarshal json: %v\n", err)
					return
				}
			} else {
				fmt.Printf("couldnt parse data structure")
				return
			}

			go func(ctx context.Context, column types.CardEvent) {
				select {
				case <-ctx.Done():
					return
				default:

					b, err := json.Marshal(cardData)
					if err != nil {
						fmt.Println("unable to serialize column data: ", err)
						return
					}

					_, err = a.repository.CreateOperation(types.CardTable, cardData.Card.ID, string(b), types.UpdateOperation)
					if err != nil {
						fmt.Println(err)
						return
					}

				}
			}(a.ctx, cardData)
		}
	})

	runtime.EventsOn(a.ctx, "card:create", func(optionalData ...interface{}) {
		if len(optionalData) > 0 {
			var cardData types.CardEvent
			data, ok := optionalData[0].(string)

			if ok {
				if err := json.Unmarshal([]byte(data), &cardData); err != nil {
					fmt.Printf("unable to unmarshal json: %v\n", err)
					return
				}
			} else {
				fmt.Printf("couldnt parse data structure")
				return
			}

			go func(ctx context.Context, payload string, card types.CardEvent) {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := a.repository.CreateOperation(types.CardTable, card.Card.ID, payload, types.InsertOperation)
					if err != nil {
						fmt.Println(err)
					}
				}
			}(a.ctx, data, cardData)
		}
	})

	runtime.EventsOn(a.ctx, "card:delete", func(optionalData ...interface{}) {
		if len(optionalData) > 0 {
			var cardData types.CardDeleteEvent
			data, ok := optionalData[0].(string)

			if ok {
				if err := json.Unmarshal([]byte(data), &cardData); err != nil {
					fmt.Printf("unable to unmarshal json: %v\n", err)
					return
				}
			} else {
				fmt.Printf("couldnt parse data structure")
				return
			}

			go func(ctx context.Context, payload string, card types.CardDeleteEvent) {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := a.repository.CreateOperation(types.CardTable, card.Card.ID, payload, types.DeleteOperation)
					if err != nil {
						fmt.Println(err)
					}
				}
			}(a.ctx, data, cardData)

		}
	})

	runtime.EventsOn(a.ctx, "card:column", func(optionalData ...interface{}) {
		if len(optionalData) > 0 {
			var cardColumnData types.CardColumnEvent
			data, ok := optionalData[0].(string)

			if ok {
				if err := json.Unmarshal([]byte(data), &cardColumnData); err != nil {
					// TODO: emit error
					fmt.Printf("unable to unmarshal json: %v\n", err)
					return
				}
			} else {
				fmt.Printf("couldnt parse data structure")
				return
			}

			go func(ctx context.Context, column types.CardColumnEvent) {
				select {
				case <-ctx.Done():
					return
				default:

					b, err := json.Marshal(cardColumnData)
					if err != nil {
						fmt.Println("unable to serialize column data: ", err)
						return
					}

					_, err = a.repository.CreateOperation(types.CardTable, cardColumnData.CardID, string(b), types.UpdateCardColumn)
					if err != nil {
						fmt.Println(err)
						return
					}

				}
			}(a.ctx, cardColumnData)

		}
	})
}
