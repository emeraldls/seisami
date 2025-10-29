package local

import (
	"database/sql"
	"seisami/app/internal/repo"
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
)

type localFuncs struct {
	repo repo.Repository
}

func NewLocalFuncs(repo repo.Repository) localFuncs {

	return localFuncs{
		repo,
	}
}

func (lf localFuncs) GetAllOperations(tableName types.TableName) ([]types.OperationSync, error) {
	ops, err := lf.repo.GetAllOperations(tableName)
	if err != nil {
		return nil, err
	}

	var operationPayload = make([]types.OperationSync, 0)

	for _, op := range ops {
		payload := types.OperationSync{
			ID:            op.ID,
			OperationType: op.OperationType,
			TableName:     op.TableName,
			RecordID:      op.RecordID,
			DeviceID:      op.DeviceID.String,
			PayloadData:   op.Payload,
			CreatedAt:     op.CreatedAt.String,
			UpdatedAt:     op.UpdatedAt.String,
		}

		operationPayload = append(operationPayload, payload)
	}

	return operationPayload, nil
}

func (lf localFuncs) UpdateLocalDB(op types.OperationSync) error {
	_ = query.Operation{
		ID:            op.ID,
		TableName:     op.TableName,
		RecordID:      op.RecordID,
		OperationType: op.OperationType,
		DeviceID:      sql.NullString{String: op.DeviceID, Valid: true},
		Payload:       op.PayloadData,
		CreatedAt:     sql.NullString{String: op.CreatedAt, Valid: true},
		UpdatedAt:     sql.NullString{String: op.UpdatedAt, Valid: true},
	}

	switch op.TableName {
	case "columns":
		switch op.OperationType {
		case "create", "update":

		}

	}

	return nil

}

func (lf localFuncs) UpdateSyncState(state query.SyncState) error {
	tableName, err := types.TableNameFromString(state.TableName)
	if err != nil {
		return err
	}
	return lf.repo.UpsertSyncState(tableName, state.LastSyncedOpID, state.LastSyncedAt)
}

func (lf localFuncs) UpsertSyncState(state query.SyncState) error {
	tableName, err := types.TableNameFromString(state.TableName)
	if err != nil {
		return err
	}
	return lf.repo.UpsertSyncState(tableName, state.LastSyncedOpID, state.LastSyncedAt)
}
