package utils

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func NullString(st string) sql.NullString {
	if st == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: st, Valid: true}
}

func NullTime(tm *time.Time) sql.NullTime {
	if tm == nil {
		return sql.NullTime{
			Valid: false,
		}
	}
	utcTime := *tm
	return sql.NullTime{Time: utcTime.In(time.Now().Location()).UTC(), Valid: true}
}

func NullUUID(u uuid.UUID) uuid.NullUUID {
	if u == uuid.Nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: u, Valid: true}
}
