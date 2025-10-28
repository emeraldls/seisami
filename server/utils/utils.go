package utils

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func StringToUUID(userID string) pgtype.UUID {
	parsed, _ := uuid.Parse(userID)
	return pgtype.UUID{Bytes: parsed, Valid: true}
}
