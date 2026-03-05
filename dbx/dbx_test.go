package dbx_test

import (
	"context"
	"database/sql"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/dbx"
)

// ---------------------------------------------------------------------------
// Error helpers
// ---------------------------------------------------------------------------

func TestIsNoRowsError(t *testing.T) {
	Convey("Given IsNoRowsError", t, func() {
		Convey("When error is sql.ErrNoRows", func() {
			So(dbx.IsNoRowsError(sql.ErrNoRows), ShouldBeTrue)
		})

		Convey("When error is a different error", func() {
			So(dbx.IsNoRowsError(sql.ErrConnDone), ShouldBeFalse)
		})

		Convey("When error is nil", func() {
			So(dbx.IsNoRowsError(nil), ShouldBeFalse)
		})
	})
}

// ---------------------------------------------------------------------------
// GetExecutor
// ---------------------------------------------------------------------------

func TestGetExecutor(t *testing.T) {
	Convey("Given GetExecutor", t, func() {
		Convey("When no transaction is in the context", func() {
			ctx := context.Background()
			// We pass nil as DB here; GetExecutor should return it as the fallback.
			// Since *sqlx.DB implements Executor, we just verify the fallback path
			// doesn't panic with a nil DB.
			exec := dbx.GetExecutor(ctx, nil)

			Convey("Then it should return the default DB (nil in this test)", func() {
				So(exec, ShouldBeNil)
			})
		})

		// Note: Testing with an actual *sqlx.Tx requires a real database connection.
		// The TransactionManager integration is best tested with an in-memory DB
		// or a dedicated integration test suite.
	})
}
