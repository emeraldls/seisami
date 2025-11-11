package utils

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func StringToUUID(userID string) pgtype.UUID {
	parsed, _ := uuid.Parse(userID)
	return pgtype.UUID{Bytes: parsed, Valid: true}
}

func ConvertTimestamptzToLocal(ts pgtype.Timestamptz) string {
	return ts.Time.Local().Format("2006-01-02 15:04:05")
}
