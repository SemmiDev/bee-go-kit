package sorting_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/sorting"
)

// ---------------------------------------------------------------------------
// QuoteFunc implementations
// ---------------------------------------------------------------------------

func TestQuoteFunctions(t *testing.T) {
	Convey("Given the column quoting functions", t, func() {
		Convey("MSSQLQuote should wrap in square brackets", func() {
			So(sorting.MSSQLQuote("NAME"), ShouldEqual, "[NAME]")
		})

		Convey("PostgresQuote should wrap in double quotes", func() {
			So(sorting.PostgresQuote("NAME"), ShouldEqual, `"NAME"`)
		})

		Convey("MySQLQuote should wrap in backticks", func() {
			So(sorting.MySQLQuote("NAME"), ShouldEqual, "`NAME`")
		})

		Convey("NoQuote should return as-is", func() {
			So(sorting.NoQuote("NAME"), ShouldEqual, "NAME")
		})
	})
}

// ---------------------------------------------------------------------------
// SortConfig
// ---------------------------------------------------------------------------

func TestNewSortConfig(t *testing.T) {
	Convey("Given NewSortConfig", t, func() {
		cfg := sorting.NewSortConfig(map[string]string{
			"Name":       "MS_USER_NAME",
			"CREATED_AT": "CREATE_DATE",
		}, "MS_USER_NAME")

		Convey("Then it should normalise keys to lowercase", func() {
			So(cfg.ValidateAndGetColumn("name"), ShouldEqual, "MS_USER_NAME")
			So(cfg.ValidateAndGetColumn("created_at"), ShouldEqual, "CREATE_DATE")
		})

		Convey("Then it should use ASC as default direction", func() {
			So(cfg.DefaultDirection, ShouldEqual, "ASC")
		})

		Convey("Then it should use MSSQLQuote by default", func() {
			So(cfg.Quote("COL"), ShouldEqual, "[COL]")
		})
	})
}

func TestSortConfig_WithQuote(t *testing.T) {
	Convey("Given a SortConfig", t, func() {
		cfg := sorting.NewSortConfig(map[string]string{"name": "NAME"}, "NAME")

		Convey("When WithQuote is called", func() {
			pgCfg := cfg.WithQuote(sorting.PostgresQuote)

			Convey("Then it should return a new config with the new QuoteFunc", func() {
				So(pgCfg.Quote("COL"), ShouldEqual, `"COL"`)
			})

			Convey("Then it should not mutate the original", func() {
				So(cfg.Quote("COL"), ShouldEqual, "[COL]")
			})
		})
	})
}

func TestSortConfig_ValidateAndGetColumn(t *testing.T) {
	Convey("Given a SortConfig with allowed columns", t, func() {
		cfg := sorting.NewSortConfig(map[string]string{
			"name": "USER_NAME",
			"age":  "USER_AGE",
		}, "USER_NAME")

		Convey("When the column is in the whitelist", func() {
			So(cfg.ValidateAndGetColumn("name"), ShouldEqual, "USER_NAME")
			So(cfg.ValidateAndGetColumn("NAME"), ShouldEqual, "USER_NAME")
		})

		Convey("When the column is not in the whitelist", func() {
			So(cfg.ValidateAndGetColumn("email"), ShouldEqual, "USER_NAME")
		})

		Convey("When the column is empty", func() {
			So(cfg.ValidateAndGetColumn(""), ShouldEqual, "USER_NAME")
		})
	})
}

func TestSortConfig_ValidateDirection(t *testing.T) {
	Convey("Given a SortConfig", t, func() {
		cfg := sorting.NewSortConfig(nil, "ID")

		Convey("When direction is DESC (case insensitive)", func() {
			So(cfg.ValidateDirection("desc"), ShouldEqual, "DESC")
			So(cfg.ValidateDirection("DESC"), ShouldEqual, "DESC")
			So(cfg.ValidateDirection("  desc  "), ShouldEqual, "DESC")
		})

		Convey("When direction is ASC or invalid", func() {
			So(cfg.ValidateDirection("asc"), ShouldEqual, "ASC")
			So(cfg.ValidateDirection("invalid"), ShouldEqual, "ASC")
			So(cfg.ValidateDirection(""), ShouldEqual, "ASC")
		})
	})
}

// ---------------------------------------------------------------------------
// Clause builders
// ---------------------------------------------------------------------------

func TestBuildOrderByClause(t *testing.T) {
	Convey("Given BuildOrderByClause", t, func() {
		cfg := sorting.NewSortConfig(map[string]string{
			"name": "MS_USER_NAME",
		}, "MS_USER_NAME")

		Convey("When config is nil", func() {
			So(sorting.BuildOrderByClause("name", "asc", nil), ShouldBeEmpty)
		})

		Convey("When valid sort is provided", func() {
			clause := sorting.BuildOrderByClause("name", "desc", cfg)
			So(clause, ShouldEqual, "[MS_USER_NAME] DESC")
		})

		Convey("When invalid column is provided", func() {
			clause := sorting.BuildOrderByClause("evil_col", "asc", cfg)
			So(clause, ShouldEqual, "[MS_USER_NAME] ASC")
		})
	})
}

func TestBuildFullOrderByClause(t *testing.T) {
	Convey("Given BuildFullOrderByClause", t, func() {
		cfg := sorting.NewSortConfig(map[string]string{
			"name": "MS_USER_NAME",
		}, "MS_USER_NAME")

		Convey("When config is nil", func() {
			So(sorting.BuildFullOrderByClause("name", "asc", nil), ShouldBeEmpty)
		})

		Convey("When valid sort is provided", func() {
			clause := sorting.BuildFullOrderByClause("name", "desc", cfg)
			So(clause, ShouldEqual, "ORDER BY [MS_USER_NAME] DESC")
		})
	})
}

func TestDateSortConfig(t *testing.T) {
	Convey("Given DateSortConfig", t, func() {
		cfg := sorting.DateSortConfig("CREATED_DATE", "MODIFIED_DATE", "ID")

		Convey("Then it should map common date aliases", func() {
			So(cfg.ValidateAndGetColumn("created_at"), ShouldEqual, "CREATED_DATE")
			So(cfg.ValidateAndGetColumn("create_date"), ShouldEqual, "CREATED_DATE")
			So(cfg.ValidateAndGetColumn("updated_at"), ShouldEqual, "MODIFIED_DATE")
			So(cfg.ValidateAndGetColumn("modify_date"), ShouldEqual, "MODIFIED_DATE")
			So(cfg.ValidateAndGetColumn("id"), ShouldEqual, "ID")
		})

		Convey("Then it should default to the ID column", func() {
			So(cfg.ValidateAndGetColumn("unknown"), ShouldEqual, "ID")
		})
	})
}
