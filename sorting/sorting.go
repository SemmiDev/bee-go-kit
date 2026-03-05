// Package sorting provides SQL-injection-safe ORDER BY clause building.
//
// It uses a whitelist-based approach: callers declare which column names are
// allowed, and the builder maps user-facing names to actual database column
// names. Any column not in the whitelist silently falls back to a default.
//
// The column quoting style is configurable via QuoteFunc so the same package
// works with MSSQL ([col]), PostgreSQL ("col"), MySQL (`col`), etc.
//
// Usage:
//
//	cfg := sorting.NewSortConfig(map[string]string{
//	    "name":       "MS_USER_NAME",
//	    "created_at": "CREATED_DATE",
//	}, "MS_USER_NAME")
//
//	clause := sorting.BuildFullOrderByClause(&filter, cfg)
//	// → "ORDER BY [MS_USER_NAME] ASC"
package sorting

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// QuoteFunc – configurable column quoting
// ---------------------------------------------------------------------------

// QuoteFunc is a function that quotes a column name for use in SQL.
// The default is MSSQLQuote which wraps in square brackets: [col].
type QuoteFunc func(column string) string

// MSSQLQuote wraps a column name in square brackets for MSSQL.
//
//	MSSQLQuote("NAME") → "[NAME]"
func MSSQLQuote(column string) string {
	return fmt.Sprintf("[%s]", column)
}

// PostgresQuote wraps a column name in double quotes for PostgreSQL.
//
//	PostgresQuote("NAME") → `"NAME"`
func PostgresQuote(column string) string {
	return fmt.Sprintf(`"%s"`, column)
}

// MySQLQuote wraps a column name in backticks for MySQL.
//
//	MySQLQuote("NAME") → "`NAME`"
func MySQLQuote(column string) string {
	return fmt.Sprintf("`%s`", column)
}

// NoQuote returns the column name as-is (no quoting).
func NoQuote(column string) string {
	return column
}

// ---------------------------------------------------------------------------
// SortConfig
// ---------------------------------------------------------------------------

// SortConfig defines the configuration for safe SQL sorting.
type SortConfig struct {
	// AllowedColumns maps user-facing column names (lowercase) to actual DB
	// column names. Only columns in this map are allowed in ORDER BY clauses.
	AllowedColumns map[string]string

	// DefaultColumn is used when no sort is specified or the column is invalid.
	DefaultColumn string

	// DefaultDirection is the default sort direction ("ASC" or "DESC").
	DefaultDirection string

	// Quote is the quoting function for column names. Defaults to MSSQLQuote.
	Quote QuoteFunc
}

// NewSortConfig creates a SortConfig with sensible defaults (ASC direction,
// MSSQL quoting). Keys in allowedColumns are normalised to lowercase.
func NewSortConfig(allowedColumns map[string]string, defaultColumn string) *SortConfig {
	// Normalise keys to lowercase for case-insensitive matching.
	normalised := make(map[string]string, len(allowedColumns))
	for k, v := range allowedColumns {
		normalised[strings.ToLower(k)] = v
	}

	return &SortConfig{
		AllowedColumns:   normalised,
		DefaultColumn:    defaultColumn,
		DefaultDirection: "ASC",
		Quote:            MSSQLQuote,
	}
}

// WithQuote returns a copy of the SortConfig with a different QuoteFunc.
func (sc *SortConfig) WithQuote(qf QuoteFunc) *SortConfig {
	cpy := *sc
	cpy.Quote = qf
	return &cpy
}

// ValidateAndGetColumn returns the actual DB column name for the requested
// user-facing column. Falls back to DefaultColumn when the column is empty or
// not in the whitelist.
func (sc *SortConfig) ValidateAndGetColumn(requestedColumn string) string {
	if requestedColumn == "" {
		return sc.DefaultColumn
	}
	if dbCol, ok := sc.AllowedColumns[strings.ToLower(requestedColumn)]; ok {
		return dbCol
	}
	return sc.DefaultColumn
}

// ValidateDirection returns a sanitised sort direction ("ASC" or "DESC").
func (sc *SortConfig) ValidateDirection(direction string) string {
	if strings.EqualFold(strings.TrimSpace(direction), "DESC") {
		return "DESC"
	}
	return "ASC"
}

// ---------------------------------------------------------------------------
// Sortable – minimal interface for sort parameters (matches pagination.Filter)
// ---------------------------------------------------------------------------

// Sortable is a minimal interface that provides sort parameters.
// The pagination.Filter type satisfies this interface.
type Sortable interface {
	// SortColumn returns the user-requested sort column name.
	SortColumn() string
	// SortDir returns the user-requested sort direction ("asc" or "desc").
	SortDir() string
}

// ---------------------------------------------------------------------------
// Clause builders
// ---------------------------------------------------------------------------

// BuildOrderByClause builds a safe ORDER BY clause (without the "ORDER BY"
// prefix) using sortBy / sortDirection strings and a SortConfig.
//
// Returns empty string if config is nil.
//
//	BuildOrderByClause("name", "desc", cfg) → "[MS_USER_NAME] DESC"
func BuildOrderByClause(sortBy, sortDirection string, config *SortConfig) string {
	if config == nil {
		return ""
	}

	column := config.ValidateAndGetColumn(sortBy)
	direction := config.ValidateDirection(sortDirection)
	quote := config.Quote
	if quote == nil {
		quote = MSSQLQuote
	}

	return fmt.Sprintf("%s %s", quote(column), direction)
}

// BuildFullOrderByClause builds a complete "ORDER BY ..." clause.
// Returns empty string if config is nil.
//
//	BuildFullOrderByClause("name", "desc", cfg) → "ORDER BY [MS_USER_NAME] DESC"
func BuildFullOrderByClause(sortBy, sortDirection string, config *SortConfig) string {
	clause := BuildOrderByClause(sortBy, sortDirection, config)
	if clause == "" {
		return ""
	}
	return "ORDER BY " + clause
}

// ---------------------------------------------------------------------------
// Pre-built configurations
// ---------------------------------------------------------------------------

// DateSortConfig creates a SortConfig for common date/ID based sorting.
// It maps "created_at", "create_date", "updated_at", "modify_date", and "id"
// to the given column names.
func DateSortConfig(createDateCol, modifyDateCol, idCol string) *SortConfig {
	return NewSortConfig(map[string]string{
		"created_at":  createDateCol,
		"create_date": createDateCol,
		"updated_at":  modifyDateCol,
		"modify_date": modifyDateCol,
		"id":          idCol,
	}, idCol)
}
