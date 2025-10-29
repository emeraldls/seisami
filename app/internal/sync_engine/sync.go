package sync_engine

import (
	"context"
	"fmt"
	"seisami/app/internal/cloud"
	"seisami/app/internal/local"
	"seisami/app/internal/repo"
	"seisami/app/types"
	"time"
)

type SyncEngine struct {
	local        local.Local
	cloud        cloud.Cloud
	repo         repo.Repository
	appCtx       context.Context
	sessionToken string
	apiUrl       string
}

func NewSyncEngine(repo repo.Repository, sessionToken string, appCtx context.Context, apiUrl string) *SyncEngine {
	local := local.NewLocalFuncs(repo)
	cloud := cloud.NewCloudFuncs(repo, sessionToken, appCtx, apiUrl)

	return &SyncEngine{
		local:        local,
		cloud:        cloud,
		appCtx:       appCtx,
		repo:         repo,
		sessionToken: sessionToken,
		apiUrl:       apiUrl,
	}
}

func (s *SyncEngine) UpdateSessionToken(token string) {
	s.sessionToken = token
	s.cloud = cloud.NewCloudFuncs(s.repo, token, s.appCtx, s.apiUrl)
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
	localOps, err := s.local.GetAllOperations(tableName)
	if err != nil {
		return err
	}

	cloudOps, err := s.cloud.GetAllOperations(tableName)
	if err != nil {
		return err
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
			operation, err := s.cloud.PullRecord(tableName)
			if err != nil {
				fmt.Printf("error pulling record from cloud: %v\n", err)
				continue
			}

			// after pulling record from cloud, you update it local db

			if err := s.local.UpdateLocalDB(operation); err != nil {
				fmt.Printf("error updating local db: %v\n", err)
				continue
			}
			pulled = true

		case !hasCloud:
			// new record exists locally, push it to cloud
			httpResp := s.cloud.PushRecord(localOp)
			if httpResp.HasError {
				fmt.Printf("push error: %v (data: %v)\n", httpResp.Message, httpResp.Data)
				continue
			}
			pushed = true

		default:
			// both exists
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
				// local record is newer, push to cloud
				httpResp := s.cloud.PushRecord(localOp)
				if httpResp.HasError {
					fmt.Printf("[Push error]: %v\n", httpResp.Message)
					continue
				}
				pushed = true

			case cloudTs > localTs:
				// cloud record is newer, pull
				operation, err := s.cloud.PullRecord(tableName)
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

	if pushed {
		syncState, err := s.repo.GetSyncState(tableName)
		if err != nil {
			fmt.Println(err)
		} else {
			if err := s.cloud.UpdateSyncState(syncState); err != nil {
				fmt.Printf("error updating cloud sync state: %v\n", err)
			}
		}

	}

	if pulled {
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
	data, err := s.repo.ExportAllData()
	if err != nil {
		return err
	}

	for _, b := range data.Boards {
		if err := s.cloud.UpsertBoard(b); err != nil {
			fmt.Printf("failed uploading board %v: %v\n", b.ID, err)
		}
	}

	for _, c := range data.Columns {
		if err := s.cloud.UpsertColumn(c); err != nil {
			fmt.Printf("failed uploading column %v: %v\n", c.ID, err)
		}
	}

	for _, card := range data.Cards {
		if err := s.cloud.UpsertCard(card); err != nil {
			fmt.Printf("failed uploading card %v: %v\n", card.ID, err)
		}
	}

	if err := s.cloud.InitializeSyncStateForUser(); err != nil {
		fmt.Printf("failed initializing cloud sync state: %v\n", err)
	}

	state, err := s.cloud.GetSyncState(types.BoardTable)
	if err == nil {
		err = s.local.UpsertSyncState(state)
		if err != nil {
			fmt.Println(err)
		}
	}

	state, err = s.cloud.GetSyncState(types.ColumnTable)
	if err == nil {
		err = s.local.UpsertSyncState(state)
		if err != nil {
			fmt.Println(err)
		}
	}

	state, err = s.cloud.GetSyncState(types.CardTable)
	if err == nil {
		err = s.local.UpsertSyncState(state)
		if err != nil {
			fmt.Println(err)
		}
	}

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
