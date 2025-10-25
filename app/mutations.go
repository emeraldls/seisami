package main

import (
	"context"
	"encoding/json"
	"fmt"
	"seisami/app/types"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) handleMutations() {
	const layout = "2006-01-02 15:04:05"

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

			go func(ctx context.Context, column types.ColumnEvent) {
				select {
				case <-ctx.Done():
					return
				default:

					b, err := json.Marshal(columnData)
					if err != nil {
						fmt.Println("unable to serialize column data: ", err)
						return
					}
					op, err := a.repository.CreateOperation(types.ColumnTable, columnData.ID, string(b), types.UpdateOperation)
					if err != nil {
						fmt.Println(err)
						return
					}

					if op.CreatedAt.Valid {
						t, err := time.Parse(layout, op.CreatedAt.String)
						if err != nil {
							fmt.Println("unable to parse created: ", err)
							return
						}

						err = a.repository.UpsertSyncState(types.ColumnTable, op.ID, t.Unix())
						if err != nil {
							fmt.Println(err)
						}

						// syncState, err := a.repository.GetSyncState("column")

						// if err != nil {

						// } else {
						// 	err = a.repository.UpdateSyncState("column", op.ID, t.Unix())
						// 	if err != nil {
						// 		fmt.Println(err)
						// 	}
						// }
					} else {
						fmt.Println("created_at is invalid")
					}

				}
			}(a.ctx, columnData)

			msg := types.Message{Action: "broadcast", RoomID: columnData.RoomID, Data: data}

			err := a.sendCollabCommand(msg)
			if err != nil {
				fmt.Printf("unable to broadcast the command: %v\n", err)
				return
			}

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
					op, err := a.repository.CreateOperation(types.CardTable, cardData.Card.ID, string(b), types.UpdateOperation)
					if err != nil {
						fmt.Println(err)
						return
					}

					if op.CreatedAt.Valid {
						t, err := time.Parse(layout, op.CreatedAt.String)
						if err != nil {
							fmt.Println("unable to parse created: ", err)
							return
						}

						err = a.repository.UpsertSyncState(types.CardTable, op.ID, t.Unix())
						if err != nil {
							fmt.Println(err)
						}

					} else {
						fmt.Println("created_at is invalid")
					}

				}
			}(a.ctx, cardData)

			msg := types.Message{Action: "broadcast", RoomID: cardData.Column.RoomID, Data: data}

			err := a.sendCollabCommand(msg)
			if err != nil {
				fmt.Printf("unable to broadcast the command: %v\n", err)
				return
			}

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
					op, err := a.repository.CreateOperation(types.CardTable, cardColumnData.CardID, string(b), types.UpdateOperation)
					if err != nil {
						fmt.Println(err)
						return
					}

					if op.CreatedAt.Valid {
						t, err := time.Parse(layout, op.CreatedAt.String)
						if err != nil {
							fmt.Println("unable to parse created: ", err)
							return
						}

						err = a.repository.UpsertSyncState(types.CardTable, op.ID, t.Unix())
						if err != nil {
							fmt.Println(err)
						}

					} else {
						fmt.Println("created_at is invalid")
					}

				}
			}(a.ctx, cardColumnData)

			msg := types.Message{Action: "broadcast", RoomID: cardColumnData.RoomID, Data: data}

			err := a.sendCollabCommand(msg)
			if err != nil {
				fmt.Printf("unable to broadcast the command: %v\n", err)
				return
			}

		}
	})
}
