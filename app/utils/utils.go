package utils

import (
	"database/sql"
	"time"
)

func ConvertTimestamptzToLocal(timeString sql.NullString) string {
	if !timeString.Valid {
		return ""
	}
	ts, err := time.Parse("2006-01-02 15:04:05", timeString.String)
	if err != nil {
		return timeString.String
	}

	return ts.Local().Format("2006-01-02 15:04:05")
}
