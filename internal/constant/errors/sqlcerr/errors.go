package sqlcerr

import (
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq"
)

var (
	ErrNoRows = pgx.ErrNoRows
)

// Is compares errors by value; used for matching sentinel errors like ErrNoRows.
func Is(err, target error) bool {
	if err == nil || target == nil {
		return false
	}
	return err.Error() == target.Error()
}

// IsDuplicate checks whether the given error represents a unique-constraint violation.
// It supports both lib/pq and pgx/pgconn error types.
func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}

	// Handle pgx/pgconn errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	// Handle lib/pq errors (for completeness / legacy)
	if duplicateError, ok := err.(*pq.Error); ok {
		return duplicateError.Code == "23505"
	}

	return false
}
