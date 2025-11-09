package sync_engine

import (
	"encoding/json"
	"fmt"
	"seisami/app/internal/cloud"
	"seisami/app/internal/local"
	"seisami/app/internal/repo"
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
	"time"
)

/*
	TODO: Cloud shouldnt be initialized inside sync engine, there are methods that are not for syncing
*/

type SyncEngine struct {
	local local.Local
	cloud cloud.Cloud
	repo  repo.Repository
}

func NewSyncEngine(repo repo.Repository, cloud cloud.Cloud) *SyncEngine {
	local := local.NewLocalFuncs(repo)

	return &SyncEngine{
		local: local,
		cloud: cloud,
		repo:  repo,
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

func (s *SyncEngine) SyncData(tableName types.TableName) error {
	fmt.Println("syncing data for table: ", tableName.String())
	localOps, err := s.local.GetAllOperations(tableName)
	if err != nil {
		return fmt.Errorf("[LOCAL] -> %v", err)
	}

	// get the local sycned state firstly
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

	cloudOps, err := s.cloud.GetAllOperations(tableName, since)
	if err != nil {
		return fmt.Errorf("[CLOUD] -> %v", err)
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
			operation, err := s.cloud.PullRecord(tableName, since)
			if err != nil {
				fmt.Printf("error pulling record from cloud: %v\n", err)
				continue
			}

			// after pulling record from cloud, you update it local db

			fmt.Println("updating local db with new record pulled")
			if err := s.local.UpdateLocalDB(operation); err != nil {
				fmt.Printf("error updating local db: %v\n", err)
				continue
			}
			pulled = true

		case !hasCloud:
			// new record exists locally, push it to cloud
			fmt.Println("new record exists locally, pushing it to cloud: ", recordId)
			httpResp := s.cloud.PushRecord(localOp)
			if httpResp.HasError {
				fmt.Printf("push error: %v (data: %v)\n", httpResp.Message, httpResp.Data)
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
				httpResp := s.cloud.PushRecord(localOp)
				if httpResp.HasError {
					fmt.Printf("[Push error]: %v\n", httpResp.Message)
					continue
				}
				pushed = true

			case cloudTs > localTs:
				fmt.Println("cloud record is newer, pulling to local")
				// cloud record is newer, pull
				operation, err := s.cloud.PullRecord(tableName, since)
				if err != nil {
					fmt.Printf("error pulling from cloud: %v\n", err)
					continue
				}
				if err := s.local.UpdateLocalDB(operation); err != nil {
					fmt.Printf("error updating local db: %v\n", err)
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
			fmt.Printf("error upserting local sync state: %v\n", err)
		}

		if err := s.cloud.UpdateSyncState(syncState); err != nil {
			fmt.Printf("error updating cloud sync state: %v\n", err)
		}
	}

	if pulled {
		fmt.Printf("\n\n<-------Pulling New Data From Cloud ----------->\n\n")
		newState, err := s.cloud.GetSyncState(tableName)
		if err != nil {
			fmt.Printf("error fetching cloud sync state: %v\n", err)
		} else {
			if err := s.local.UpdateSyncState(newState); err != nil {
				fmt.Printf("error updating local sync state: %v\n", err)
			}
		}
	}

	return nil
}

func (s *SyncEngine) BootstrapCloud() error {

	localData, err := s.repo.ExportAllData()
	if err != nil {
		return fmt.Errorf("failed to export local data: %v", err)
	}

	cloudData, err := s.cloud.ImportAllUserData()
	if err != nil {
		return fmt.Errorf("failed to import cloud data: %v", err)
	}

	mergedBoards := s.mergeBoards(localData.Boards, cloudData.Boards)

	for _, b := range mergedBoards {
		if err := s.cloud.UpsertBoard(b); err != nil {
			fmt.Printf("failed uploading board %v: %v\n", b.ID, err)
		}
		if _, err := s.repo.ImportBoard(b.ID, b.Name, b.CreatedAt, b.UpdatedAt); err != nil {
			fmt.Printf("failed importing board locally %v: %v\n", b.ID, err)
		}
	}

	mergedColumns := s.mergeColumns(localData.Columns, cloudData.Columns)
	for _, c := range mergedColumns {
		if err := s.cloud.UpsertColumn(c); err != nil {
			fmt.Printf("failed uploading column %v: %v\n", c.ID, err)
		}
		if _, err := s.repo.ImportColumn(c.ID, c.BoardID, c.Name, c.Position, c.CreatedAt, c.UpdatedAt); err != nil {
			fmt.Printf("failed importing column locally %v: %v\n", c.ID, err)
		}
	}

	mergedCards := s.mergeCards(localData.Cards, cloudData.Cards)
	for _, card := range mergedCards {
		if err := s.cloud.UpsertCard(card); err != nil {
			fmt.Printf("failed uploading card %v: %v\n", card.ID, err)
		}
		if _, err := s.repo.ImportCard(card.ID, card.ColumnID, card.Title, card.Description, card.Attachments, card.CreatedAt, card.UpdatedAt); err != nil {
			fmt.Printf("failed importing card locally %v: %v\n", card.ID, err)
		}
	}

	mergedTranscriptions := s.mergeTranscriptions(localData.Transcriptions, cloudData.Transcriptions)
	for _, t := range mergedTranscriptions {
		// Upload to cloud (you'll need to add this method)
		// if err := s.cloud.UpsertTranscription(t); err != nil {
		//     fmt.Printf("failed uploading transcription %v: %v\n", t.ID, err)
		// }
		if _, err := s.repo.ImportTranscription(t.ID, t.BoardID, t.Transcription, t.RecordingPath, t.Intent, t.AssistantResponse, t.CreatedAt, t.UpdatedAt); err != nil {
			fmt.Printf("failed importing transcription locally %v: %v\n", t.ID, err)
		}
	}

	if err := s.cloud.InitializeSyncStateForUser(); err != nil {
		fmt.Printf("failed initializing cloud sync state: %v\n", err)
	}

	for _, table := range []types.TableName{types.BoardTable, types.ColumnTable, types.CardTable} {
		state, err := s.cloud.GetSyncState(table)
		if err == nil {
			if err = s.local.UpsertSyncState(state); err != nil {
				fmt.Printf("error syncing state for %s: %v\n", table, err)
			}
		} else {
			fmt.Printf("get sync state error for %s: %v\n", table, err)
		}
	}

	fmt.Println("Bootstrap completed successfully - all data merged and synced")
	return nil
}

func (s *SyncEngine) ImportNewBoard(boardID string) error {
	boardData, err := s.cloud.ImportBoardData(boardID)
	if err != nil {
		return fmt.Errorf("failed to fetch board data from cloud: %v", err)
	}

	bByte, _ := json.MarshalIndent(boardData, "", " ")
	fmt.Println(string(bByte))

	existingBoard, err := s.repo.GetBoard(boardData.Board.ID)
	if err == nil && existingBoard.ID != "" {
		return fmt.Errorf("board '%s' already exists locally", boardData.Board.Name)
	}

	_, err = s.repo.ImportBoard(
		boardData.Board.ID,
		boardData.Board.Name,
		boardData.Board.CreatedAt,
		boardData.Board.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to import board: %v", err)
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
			fmt.Printf("failed to import column %s: %v\n", col.ID, err)
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
			fmt.Printf("failed to import card %s: %v\n", card.ID, err)
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
			fmt.Printf("failed to import transcription %s: %v\n", t.ID, err)
			continue
		}
	}

	fmt.Printf("Successfully imported board %s with %d columns, %d cards, and %d transcriptions\n",
		boardData.Board.ID,
		len(boardData.Columns),
		len(boardData.Cards),
		len(boardData.Transcriptions),
	)

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
