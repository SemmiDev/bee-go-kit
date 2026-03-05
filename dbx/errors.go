package dbx

import (
	"database/sql"
	"errors"

	mssql "github.com/microsoft/go-mssqldb"
)

// ---------------------------------------------------------------------------
// Database error helpers
// ---------------------------------------------------------------------------

// IsUniqueConstraintError reports whether err is a MSSQL unique-constraint
// violation (error numbers 2627 or 2601).
//
// This is MSSQL-specific. For PostgreSQL you would check for error code 23505.
func IsUniqueConstraintError(err error) bool {
	var mssqlErr mssql.Error
	if errors.As(err, &mssqlErr) {
		return mssqlErr.Number == 2627 || mssqlErr.Number == 2601
	}
	return false
}

// IsNoRowsError reports whether err is sql.ErrNoRows, indicating that a
// query returned zero rows.
func IsNoRowsError(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
