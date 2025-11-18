package sync_engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"seisami/app/internal/cloud"
	"seisami/app/internal/local"
	"seisami/app/internal/repo"
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

/*
	TODO: Cloud shouldnt be initialized inside sync engine, there are methods that are not for syncing
*/

type SyncEngine struct {
	local local.Local
	cloud cloud.Cloud
	repo  repo.Repository
	ctx   context.Context
}

func NewSyncEngine(repo repo.Repository, cloud cloud.Cloud, ctx context.Context) *SyncEngine {
	local := local.NewLocalFuncs(repo)

	return &SyncEngine{
		local: local,
		cloud: cloud,
		repo:  repo,
		ctx:   ctx,
	}
}

func (s *SyncEngine) emitError(event string, message string) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, event, message)
	}
}

func (s *SyncEngine) emitSuccess(event string, data interface{}) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, event, data)
	}
}

const layout = "2006-01-02 15:04:05"

/*
if i pass in alot of ops, then for a particular record,
it has a create operation initially, then a user updates the same record & delete eventually,
is it all that will be returned? because imagine we see an update type,
we're to run sql update, then what if the data we're updating locally doesnt exist?

----
Something is wrong
*/

/*
	This reason why we have pull record is for when you're team,
*/

func (s *SyncEngine) SyncData(tableName types.TableName, silent bool) error {
	fmt.Println("syncing data for table: ", tableName.String())
	if !silent {
		s.emitSuccess("sync:started", map[string]string{"table": tableName.String()})
	}

	localOps, err := s.local.GetAllOperations(tableName)
	if err != nil {
		errMsg := fmt.Sprintf("[LOCAL] failed to get operations: %v", err)
		if !silent {
			s.emitError("sync:error", errMsg)
		}
		return errors.New(errMsg)
	}

	// get the local synced state firstly
	var since int64 = 0
	syncState, err := s.repo.GetSyncState(tableName)
	if err != nil {
		s.local.UpsertSyncState(query.SyncState{
			TableName:    tableName.String(),
			LastSyncedAt: 0,
		})
	}

	if syncState.LastSyncedAt != 0 {
		since = syncState.LastSyncedAt
	}

	cloudResp := s.cloud.GetAllOperations(tableName, since)
	if cloudResp.Error != "" {
		errMsg := fmt.Sprintf("[CLOUD] failed to get operations: %s", cloudResp.Error)
		if !silent {
			s.emitError("sync:error", errMsg)
		}
		return errors.New(errMsg)
	}

	cloudOps, ok := cloudResp.Data.([]types.OperationSync)
	if !ok {
		errMsg := "[CLOUD] invalid response data type"
		if !silent {
			s.emitError("sync:error", errMsg)
		}
		return errors.New(errMsg)
	}

	localLatest := latestByRecord(localOps)
	cloudLatest := latestByRecord(cloudOps)

	var pushed, pulled bool

	for recordId := range unionKeys(localLatest, cloudLatest) {
		localOp, hasLocal := localLatest[recordId]
		cloudOp, hasCloud := cloudLatest[recordId]

		switch {
		case !hasLocal:
			// new record exists in cloud, pull it
			fmt.Println("new record exists in cloud, pulling it: ", recordId)
			pullResp := s.cloud.PullRecord(tableName, since)
			if pullResp.Error != "" {
				errMsg := fmt.Sprintf("error pulling record from cloud: %s", pullResp.Error)
				fmt.Println(errMsg)
				if !silent {
					s.emitError("sync:pull_error", errMsg)
				}
				continue
			}

			operation, ok := pullResp.Data.(types.OperationSync)
			if !ok {
				errMsg := "invalid pull response data type"
				fmt.Println(errMsg)
				if !silent {
					s.emitError("sync:pull_error", errMsg)
				}
				continue
			}

			// after pulling record from cloud, you update it local db
			fmt.Println("updating local db with new record pulled")
			if err := s.local.UpdateLocalDB(operation); err != nil {
				errMsg := fmt.Sprintf("error updating local db: %v", err)
				fmt.Println(errMsg)
				if !silent {
					s.emitError("sync:local_update_error", errMsg)
				}
				continue
			}
			pulled = true

		case !hasCloud:
			// new record exists locally, push it to cloud
			fmt.Println("new record exists locally, pushing it to cloud: ", recordId)
			pushResp := s.cloud.PushRecord(localOp)
			if pushResp.Error != "" {
				errMsg := fmt.Sprintf("push error: %s", pushResp.Error)
				fmt.Println(errMsg)
				if !silent {
					s.emitError("sync:push_error", errMsg)
				}
				continue
			}
			pushed = true

		default:
			// both exists
			fmt.Println("both records exist, comparing timestamps")
			localTime, err := time.Parse(layout, localOp.CreatedAt)
			if err != nil {
				fmt.Printf("error parsing local createdAt: %v\n", err)
				continue
			}

			cloudTime, err := time.Parse(layout, cloudOp.CreatedAt)
			if err != nil {
				fmt.Printf("error parsing cloud createdAt: %v\n", err)
				continue
			}

			localTs := localTime.Unix()
			cloudTs := cloudTime.Unix()

			switch {
			case localTs > cloudTs:
				fmt.Println("local record is newer, pushing to cloud")
				// local record is newer, push to cloud
				pushResp := s.cloud.PushRecord(localOp)
				if pushResp.Error != "" {
					errMsg := fmt.Sprintf("[Push error]: %s", pushResp.Error)
					fmt.Println(errMsg)
					if !silent {
						s.emitError("sync:push_error", errMsg)
					}
					continue
				}
				pushed = true

			case cloudTs > localTs:
				fmt.Println("cloud record is newer, pulling to local")
				// cloud record is newer, pull
				pullResp := s.cloud.PullRecord(tableName, since)
				if pullResp.Error != "" {
					errMsg := fmt.Sprintf("error pulling from cloud: %s", pullResp.Error)
					fmt.Println(errMsg)
					if !silent {
						s.emitError("sync:pull_error", errMsg)
					}
					continue
				}

				operation, ok := pullResp.Data.(types.OperationSync)
				if !ok {
					errMsg := "invalid pull response data type"
					fmt.Println(errMsg)
					if !silent {
						s.emitError("sync:pull_error", errMsg)
					}
					continue
				}

				if err := s.local.UpdateLocalDB(operation); err != nil {
					errMsg := fmt.Sprintf("error updating local db: %v", err)
					fmt.Println(errMsg)
					if !silent {
						s.emitError("sync:local_update_error", errMsg)
					}
					continue
				}
				pulled = true

			default:

			}
		}
	}

	fmt.Printf("Pushed State: %v, Pulled State: %v\n", pushed, pulled)

	if pushed {
		fmt.Printf("\n\n<-------Pushing New Data To Cloud ----------->\n\n")

		lastOpID := ""
		if len(localOps) > 0 {

			var latestOp types.OperationSync
			var latestTime int64

			for _, op := range localOps {
				t, err := time.Parse(layout, op.CreatedAt)
				if err != nil {
					continue
				}
				ts := t.Unix()
				if ts > latestTime {
					latestTime = ts
					latestOp = op
				}
			}
			lastOpID = latestOp.ID
		}

		syncState := query.SyncState{
			TableName:      tableName.String(),
			LastSyncedAt:   time.Now().Unix(),
			LastSyncedOpID: lastOpID,
		}

		if err := s.local.UpsertSyncState(syncState); err != nil {
			errMsg := fmt.Sprintf("error upserting local sync state: %v", err)
			fmt.Println(errMsg)
			if !silent {
				s.emitError("sync:state_error", errMsg)
			}
		}

		updateResp := s.cloud.UpdateSyncState(syncState)
		if updateResp.Error != "" {
			errMsg := fmt.Sprintf("error updating cloud sync state: %s", updateResp.Error)
			fmt.Println(errMsg)
			if !silent {
				s.emitError("sync:state_error", errMsg)
			}
		}
	}

	if pulled {
		fmt.Printf("\n\n<-------Pulling New Data From Cloud ----------->\n\n")
		stateResp := s.cloud.GetSyncState(tableName)
		if stateResp.Error != "" {
			errMsg := fmt.Sprintf("error fetching cloud sync state: %s", stateResp.Error)
			fmt.Println(errMsg)
			if !silent {
				s.emitError("sync:state_error", errMsg)
			}
		} else {
			newState, ok := stateResp.Data.(query.SyncState)
			if !ok {
				errMsg := "invalid sync state response data type"
				fmt.Println(errMsg)
				if !silent {
					s.emitError("sync:state_error", errMsg)
				}
			} else {
				if err := s.local.UpdateSyncState(newState); err != nil {
					errMsg := fmt.Sprintf("error updating local sync state: %v", err)
					fmt.Println(errMsg)
					if !silent {
						s.emitError("sync:state_error", errMsg)
					}
				}
			}
		}
	}

	if !silent {
		s.emitSuccess("sync:completed", map[string]interface{}{
			"table":  tableName.String(),
			"pushed": pushed,
			"pulled": pulled,
		})
	}

	return nil
}

func (s *SyncEngine) BootstrapCloud() error {
	s.emitSuccess("bootstrap:started", "Starting cloud bootstrap")

	localData, err := s.repo.ExportAllData()
	if err != nil {
		errMsg := fmt.Sprintf("failed to export local data: %v", err)
		s.emitError("bootstrap:error", errMsg)
		return errors.New(errMsg)
	}

	cloudResp := s.cloud.ImportAllUserData()
	if cloudResp.Error != "" {
		errMsg := fmt.Sprintf("failed to import cloud data: %s", cloudResp.Error)
		s.emitError("bootstrap:error", errMsg)
		return errors.New(errMsg)
	}

	cloudData, ok := cloudResp.Data.(types.ExportedData)
	if !ok {
		errMsg := "invalid cloud data response type"
		s.emitError("bootstrap:error", errMsg)
		return errors.New(errMsg)
	}

	mergedBoards := s.mergeBoards(localData.Boards, cloudData.Boards)

	for _, b := range mergedBoards {
		upsertResp := s.cloud.UpsertBoard(b)
		if upsertResp.Error != "" {
			errMsg := fmt.Sprintf("failed uploading board %v: %s", b.ID, upsertResp.Error)
			fmt.Println(errMsg)
			s.emitError("bootstrap:board_error", errMsg)
		}
		if _, err := s.repo.ImportBoard(b.ID, b.Name, b.CreatedAt, b.UpdatedAt); err != nil {
			errMsg := fmt.Sprintf("failed importing board locally %v: %v", b.ID, err)
			fmt.Println(errMsg)
			s.emitError("bootstrap:board_error", errMsg)
		}
	}

	mergedColumns := s.mergeColumns(localData.Columns, cloudData.Columns)
	for _, c := range mergedColumns {
		upsertResp := s.cloud.UpsertColumn(c)
		if upsertResp.Error != "" {
			errMsg := fmt.Sprintf("failed uploading column %v: %s", c.ID, upsertResp.Error)
			fmt.Println(errMsg)
			s.emitError("bootstrap:column_error", errMsg)
		}
		if _, err := s.repo.ImportColumn(c.ID, c.BoardID, c.Name, c.Position, c.CreatedAt, c.UpdatedAt); err != nil {
			errMsg := fmt.Sprintf("failed importing column locally %v: %v", c.ID, err)
			fmt.Println(errMsg)
			s.emitError("bootstrap:column_error", errMsg)
		}
	}

	mergedCards := s.mergeCards(localData.Cards, cloudData.Cards)
	for _, card := range mergedCards {
		upsertResp := s.cloud.UpsertCard(card)
		if upsertResp.Error != "" {
			errMsg := fmt.Sprintf("failed uploading card %v: %s", card.ID, upsertResp.Error)
			fmt.Println(errMsg)
			s.emitError("bootstrap:card_error", errMsg)
		}
		if _, err := s.repo.ImportCard(card.ID, card.ColumnID, card.Title, card.Description, card.Attachments, card.CreatedAt, card.UpdatedAt); err != nil {
			errMsg := fmt.Sprintf("failed importing card locally %v: %v", card.ID, err)
			fmt.Println(errMsg)
			s.emitError("bootstrap:card_error", errMsg)
		}
	}

	mergedTranscriptions := s.mergeTranscriptions(localData.Transcriptions, cloudData.Transcriptions)
	for _, t := range mergedTranscriptions {
		// Upload to cloud (you'll need to add this method)
		// if err := s.cloud.UpsertTranscription(t); err != nil {
		//     fmt.Printf("failed uploading transcription %v: %v\n", t.ID, err)
		// }
		if _, err := s.repo.ImportTranscription(t.ID, t.BoardID, t.Transcription, t.RecordingPath, t.Intent, t.AssistantResponse, t.CreatedAt, t.UpdatedAt); err != nil {
			errMsg := fmt.Sprintf("failed importing transcription locally %v: %v", t.ID, err)
			fmt.Println(errMsg)
			s.emitError("bootstrap:transcription_error", errMsg)
		}
	}

	initResp := s.cloud.InitializeSyncStateForUser()
	if initResp.Error != "" {
		errMsg := fmt.Sprintf("failed initializing cloud sync state: %s", initResp.Error)
		fmt.Println(errMsg)
		s.emitError("bootstrap:init_error", errMsg)
	}

	for _, table := range []types.TableName{types.BoardTable, types.ColumnTable, types.CardTable} {
		stateResp := s.cloud.GetSyncState(table)
		if stateResp.Error != "" {
			errMsg := fmt.Sprintf("get sync state error for %s: %s", table, stateResp.Error)
			fmt.Println(errMsg)
			s.emitError("bootstrap:state_error", errMsg)
		} else {
			state, ok := stateResp.Data.(query.SyncState)
			if !ok {
				errMsg := fmt.Sprintf("invalid sync state response for %s", table)
				fmt.Println(errMsg)
				s.emitError("bootstrap:state_error", errMsg)
				continue
			}

			if err := s.local.UpsertSyncState(state); err != nil {
				errMsg := fmt.Sprintf("error syncing state for %s: %v", table, err)
				fmt.Println(errMsg)
				s.emitError("bootstrap:state_error", errMsg)
			}
		}
	}

	fmt.Println("Bootstrap completed successfully - all data merged and synced")
	s.emitSuccess("bootstrap:completed", "Bootstrap completed successfully")
	return nil
}

func (s *SyncEngine) ImportNewBoard(boardID string) error {
	s.emitSuccess("import:started", map[string]string{"boardId": boardID})

	boardResp := s.cloud.ImportBoardData(boardID)
	if boardResp.Error != "" {
		errMsg := fmt.Sprintf("failed to fetch board data from cloud: %s", boardResp.Error)
		s.emitError("import:error", errMsg)
		return errors.New(errMsg)
	}

	boardData, ok := boardResp.Data.(types.ImportUserBoardData)
	if !ok {
		errMsg := "invalid board data response type"
		s.emitError("import:error", errMsg)
		return errors.New(errMsg)
	}

	bByte, _ := json.MarshalIndent(boardData, "", " ")
	fmt.Println(string(bByte))

	existingBoard, err := s.repo.GetBoard(boardData.Board.ID)
	if err == nil && existingBoard.ID != "" {
		errMsg := fmt.Sprintf("board '%s' already exists locally", boardData.Board.Name)
		s.emitError("import:error", errMsg)
		return errors.New(errMsg)
	}

	_, err = s.repo.ImportBoard(
		boardData.Board.ID,
		boardData.Board.Name,
		boardData.Board.CreatedAt,
		boardData.Board.UpdatedAt,
	)
	if err != nil {
		errMsg := fmt.Sprintf("failed to import board: %v", err)
		s.emitError("import:error", errMsg)
		return errors.New(errMsg)
	}

	for _, col := range boardData.Columns {
		_, err = s.repo.ImportColumn(
			col.ID,
			col.BoardID,
			col.Name,
			col.Position,
			col.CreatedAt,
			col.UpdatedAt,
		)
		if err != nil {
			errMsg := fmt.Sprintf("failed to import column %s: %v", col.ID, err)
			fmt.Println(errMsg)
			s.emitError("import:column_error", errMsg)
			continue
		}
	}

	for _, card := range boardData.Cards {
		_, err = s.repo.ImportCard(
			card.ID,
			card.ColumnID,
			card.Title,
			card.Description,
			card.Attachments,
			card.CreatedAt,
			card.UpdatedAt,
		)
		if err != nil {
			errMsg := fmt.Sprintf("failed to import card %s: %v", card.ID, err)
			fmt.Println(errMsg)
			s.emitError("import:card_error", errMsg)
			continue
		}
	}

	for _, t := range boardData.Transcriptions {
		_, err = s.repo.ImportTranscription(
			t.ID,
			t.BoardID,
			t.Transcription,
			t.RecordingPath,
			t.Intent,
			t.AssistantResponse,
			t.CreatedAt,
			t.UpdatedAt,
		)
		if err != nil {
			errMsg := fmt.Sprintf("failed to import transcription %s: %v", t.ID, err)
			fmt.Println(errMsg)
			s.emitError("import:transcription_error", errMsg)
			continue
		}
	}

	successMsg := fmt.Sprintf("Successfully imported board %s with %d columns, %d cards, and %d transcriptions",
		boardData.Board.ID,
		len(boardData.Columns),
		len(boardData.Cards),
		len(boardData.Transcriptions),
	)
	fmt.Println(successMsg)
	s.emitSuccess("import:completed", map[string]interface{}{
		"boardId":        boardData.Board.ID,
		"boardName":      boardData.Board.Name,
		"columnsCount":   len(boardData.Columns),
		"cardsCount":     len(boardData.Cards),
		"transcriptions": len(boardData.Transcriptions),
	})

	return nil
}

func latestByRecord(ops []types.OperationSync) map[string]types.OperationSync {
	m := make(map[string]types.OperationSync)
	for _, op := range ops {

		t, err := time.Parse(layout, op.CreatedAt)
		if err != nil {
			fmt.Printf("unable to parse created_at to time layout: %v", err)
			continue
		}

		createdAt := t.Unix()

		existing, ok := m[op.RecordID]

		if !ok {
			m[op.RecordID] = op
			continue
		}

		existingTime, err := time.Parse(layout, existing.CreatedAt)
		if err != nil {
			fmt.Printf("unable to parse existing created_at into time layout: %v", err)
			continue
		}

		existingAt := existingTime.Unix()

		if createdAt > existingAt {
			m[op.RecordID] = op
		}
	}

	return m
}

func unionKeys(a, b map[string]types.OperationSync) map[string]struct{} {
	keys := make(map[string]struct{})
	for k := range a {
		keys[k] = struct{}{}
	}

	for k := range b {
		keys[k] = struct{}{}
	}

	return keys
}

func (s *SyncEngine) mergeBoards(local, cloud []types.ExportedBoard) []types.ExportedBoard {
	boardMap := make(map[string]types.ExportedBoard)

	for _, b := range local {
		boardMap[b.ID] = b
	}

	for _, cloudBoard := range cloud {
		localBoard, exists := boardMap[cloudBoard.ID]
		if !exists {
			boardMap[cloudBoard.ID] = cloudBoard
			continue
		}

		localTime, err1 := time.Parse(layout, localBoard.UpdatedAt)
		cloudTime, err2 := time.Parse(layout, cloudBoard.UpdatedAt)

		if err1 != nil || err2 != nil {
			boardMap[cloudBoard.ID] = cloudBoard
			continue
		}

		if cloudTime.After(localTime) {
			boardMap[cloudBoard.ID] = cloudBoard
		}
		// else leave local version (already in map)
	}

	result := make([]types.ExportedBoard, 0, len(boardMap))
	for _, b := range boardMap {
		result = append(result, b)
	}
	return result
}

func (s *SyncEngine) mergeColumns(local, cloud []types.ExportedColumn) []types.ExportedColumn {
	columnMap := make(map[string]types.ExportedColumn)

	for _, c := range local {
		columnMap[c.ID] = c
	}

	for _, cloudCol := range cloud {
		localCol, exists := columnMap[cloudCol.ID]
		if !exists {
			columnMap[cloudCol.ID] = cloudCol
			continue
		}

		localTime, err1 := time.Parse(layout, localCol.UpdatedAt)
		cloudTime, err2 := time.Parse(layout, cloudCol.UpdatedAt)

		if err1 != nil || err2 != nil {
			columnMap[cloudCol.ID] = cloudCol
			continue
		}

		if cloudTime.After(localTime) {
			columnMap[cloudCol.ID] = cloudCol
		}
	}

	result := make([]types.ExportedColumn, 0, len(columnMap))
	for _, c := range columnMap {
		result = append(result, c)
	}
	return result
}

func (s *SyncEngine) mergeCards(local, cloud []types.ExportedCard) []types.ExportedCard {
	cardMap := make(map[string]types.ExportedCard)

	for _, c := range local {
		cardMap[c.ID] = c
	}

	for _, cloudCard := range cloud {
		localCard, exists := cardMap[cloudCard.ID]
		if !exists {
			cardMap[cloudCard.ID] = cloudCard
			continue
		}

		localTime, err1 := time.Parse(layout, localCard.UpdatedAt)
		cloudTime, err2 := time.Parse(layout, cloudCard.UpdatedAt)

		if err1 != nil || err2 != nil {
			cardMap[cloudCard.ID] = cloudCard
			continue
		}

		if cloudTime.After(localTime) {
			cardMap[cloudCard.ID] = cloudCard
		}
	}

	result := make([]types.ExportedCard, 0, len(cardMap))
	for _, c := range cardMap {
		result = append(result, c)
	}
	return result
}

func (s *SyncEngine) mergeTranscriptions(local, cloud []types.ExportedTranscription) []types.ExportedTranscription {
	transcriptionMap := make(map[string]types.ExportedTranscription)

	for _, t := range local {
		transcriptionMap[t.ID] = t
	}

	for _, cloudTrans := range cloud {
		localTrans, exists := transcriptionMap[cloudTrans.ID]
		if !exists {
			transcriptionMap[cloudTrans.ID] = cloudTrans
			continue
		}

		localTime, err1 := time.Parse(layout, localTrans.UpdatedAt)
		cloudTime, err2 := time.Parse(layout, cloudTrans.UpdatedAt)

		if err1 != nil || err2 != nil {
			transcriptionMap[cloudTrans.ID] = cloudTrans
			continue
		}

		if cloudTime.After(localTime) {
			transcriptionMap[cloudTrans.ID] = cloudTrans
		}
	}

	result := make([]types.ExportedTranscription, 0, len(transcriptionMap))
	for _, t := range transcriptionMap {
		result = append(result, t)
	}
	return result
}
