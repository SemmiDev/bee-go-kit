package timeutil_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/timeutil"
)

// referenceTime creates a fixed time for deterministic tests.
func referenceTime() time.Time {
	return time.Date(2026, time.March, 5, 14, 1, 30, 0, time.UTC)
}

// ---------------------------------------------------------------------------
// Indonesian formatting
// ---------------------------------------------------------------------------

func TestFormatIndonesian(t *testing.T) {
	Convey("Given FormatIndonesian", t, func() {
		Convey("When a UTC time is provided", func() {
			// 14:01 UTC → 21:01 WIB (UTC+7)
			result := timeutil.FormatIndonesian(referenceTime())

			Convey("Then it should format in Indonesian with WIB timezone", func() {
				So(result, ShouldContainSubstring, "Maret")
				So(result, ShouldContainSubstring, "2026")
				So(result, ShouldContainSubstring, "WIB")
			})
		})
	})
}

func TestFormatIndonesianFull(t *testing.T) {
	Convey("Given FormatIndonesianFull", t, func() {
		result := timeutil.FormatIndonesianFull(referenceTime())

		Convey("Then it should include the day name in Indonesian", func() {
			So(result, ShouldContainSubstring, "Maret")
			So(result, ShouldContainSubstring, "WIB")
			// The day name should be one of the Indonesian day names
			dayNames := []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"}
			found := false
			for _, d := range dayNames {
				if result[:len(d)] == d || len(result) > len(d) {
					// Just check it contains a day name
				}
				if contains(result, d) {
					found = true
					break
				}
			}
			So(found, ShouldBeTrue)
		})
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFormatIndonesianDate(t *testing.T) {
	Convey("Given FormatIndonesianDate", t, func() {
		result := timeutil.FormatIndonesianDate(referenceTime())

		Convey("Then it should format as '5 Maret 2026'", func() {
			So(result, ShouldEqual, "5 Maret 2026")
		})
	})
}

func TestMonthName(t *testing.T) {
	Convey("Given MonthName", t, func() {
		So(timeutil.MonthName(time.January), ShouldEqual, "Januari")
		So(timeutil.MonthName(time.May), ShouldEqual, "Mei")
		So(timeutil.MonthName(time.December), ShouldEqual, "Desember")
	})
}

func TestDayName(t *testing.T) {
	Convey("Given DayName", t, func() {
		So(timeutil.DayName(time.Monday), ShouldEqual, "Senin")
		So(timeutil.DayName(time.Friday), ShouldEqual, "Jumat")
		So(timeutil.DayName(time.Sunday), ShouldEqual, "Minggu")
	})
}

// ---------------------------------------------------------------------------
// Standard formatting
// ---------------------------------------------------------------------------

func TestFormatDate(t *testing.T) {
	Convey("Given FormatDate", t, func() {
		result := timeutil.FormatDate(referenceTime())
		So(result, ShouldEqual, "2026-03-05")
	})
}

func TestFormatDateTime(t *testing.T) {
	Convey("Given FormatDateTime", t, func() {
		result := timeutil.FormatDateTime(referenceTime())
		So(result, ShouldEqual, "2026-03-05 14:01:30")
	})
}

func TestFormatRFC3339(t *testing.T) {
	Convey("Given FormatRFC3339", t, func() {
		result := timeutil.FormatRFC3339(referenceTime())
		So(result, ShouldContainSubstring, "2026-03-05T14:01:30")
	})
}

// ---------------------------------------------------------------------------
// Parsing
// ---------------------------------------------------------------------------

func TestParseDate(t *testing.T) {
	Convey("Given ParseDate", t, func() {
		Convey("When input is valid", func() {
			parsed, err := timeutil.ParseDate("2026-03-05")
			So(err, ShouldBeNil)
			So(parsed.Year(), ShouldEqual, 2026)
			So(parsed.Month(), ShouldEqual, time.March)
			So(parsed.Day(), ShouldEqual, 5)
		})

		Convey("When input is invalid", func() {
			_, err := timeutil.ParseDate("not-a-date")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestParseDateTime(t *testing.T) {
	Convey("Given ParseDateTime", t, func() {
		Convey("When input is valid", func() {
			parsed, err := timeutil.ParseDateTime("2026-03-05 14:01:30")
			So(err, ShouldBeNil)
			So(parsed.Hour(), ShouldEqual, 14)
			So(parsed.Minute(), ShouldEqual, 1)
		})

		Convey("When input is invalid", func() {
			_, err := timeutil.ParseDateTime("bad")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestMustParseDate(t *testing.T) {
	Convey("Given MustParseDate", t, func() {
		Convey("When input is valid", func() {
			parsed := timeutil.MustParseDate("2026-01-15")
			So(parsed.Year(), ShouldEqual, 2026)
		})

		Convey("When input is invalid", func() {
			Convey("Then it should panic", func() {
				So(func() { timeutil.MustParseDate("bad") }, ShouldPanic)
			})
		})
	})
}

// ---------------------------------------------------------------------------
// Boundaries
// ---------------------------------------------------------------------------

func TestStartOfDay(t *testing.T) {
	Convey("Given StartOfDay", t, func() {
		result := timeutil.StartOfDay(referenceTime())

		Convey("Then time should be 00:00:00.000", func() {
			So(result.Hour(), ShouldEqual, 0)
			So(result.Minute(), ShouldEqual, 0)
			So(result.Second(), ShouldEqual, 0)
			So(result.Nanosecond(), ShouldEqual, 0)
		})

		Convey("Then date should be preserved", func() {
			So(result.Year(), ShouldEqual, 2026)
			So(result.Month(), ShouldEqual, time.March)
			So(result.Day(), ShouldEqual, 5)
		})
	})
}

func TestEndOfDay(t *testing.T) {
	Convey("Given EndOfDay", t, func() {
		result := timeutil.EndOfDay(referenceTime())

		Convey("Then time should be 23:59:59", func() {
			So(result.Hour(), ShouldEqual, 23)
			So(result.Minute(), ShouldEqual, 59)
			So(result.Second(), ShouldEqual, 59)
		})
	})
}

func TestStartOfMonth(t *testing.T) {
	Convey("Given StartOfMonth", t, func() {
		result := timeutil.StartOfMonth(referenceTime())
		So(result.Day(), ShouldEqual, 1)
		So(result.Hour(), ShouldEqual, 0)
	})
}

func TestEndOfMonth(t *testing.T) {
	Convey("Given EndOfMonth", t, func() {
		Convey("When month is March (31 days)", func() {
			result := timeutil.EndOfMonth(referenceTime())
			So(result.Day(), ShouldEqual, 31)
		})

		Convey("When month is February in non-leap year", func() {
			feb := time.Date(2025, time.February, 15, 0, 0, 0, 0, time.UTC)
			result := timeutil.EndOfMonth(feb)
			So(result.Day(), ShouldEqual, 28)
		})
	})
}

// ---------------------------------------------------------------------------
// Comparison
// ---------------------------------------------------------------------------

func TestIsSameDay(t *testing.T) {
	Convey("Given IsSameDay", t, func() {
		Convey("When times are on the same day", func() {
			a := time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)
			b := time.Date(2026, 3, 5, 23, 59, 59, 0, time.UTC)
			So(timeutil.IsSameDay(a, b), ShouldBeTrue)
		})

		Convey("When times are on different days", func() {
			a := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
			b := time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC)
			So(timeutil.IsSameDay(a, b), ShouldBeFalse)
		})
	})
}

func TestDaysBetween(t *testing.T) {
	Convey("Given DaysBetween", t, func() {
		Convey("When dates are 3 days apart", func() {
			a := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
			b := time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC)
			So(timeutil.DaysBetween(a, b), ShouldEqual, 3)
		})

		Convey("When dates are reversed", func() {
			a := time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC)
			b := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
			So(timeutil.DaysBetween(a, b), ShouldEqual, 3)
		})

		Convey("When same day", func() {
			a := time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)
			So(timeutil.DaysBetween(a, a), ShouldEqual, 0)
		})
	})
}

// ---------------------------------------------------------------------------
// Convenience
// ---------------------------------------------------------------------------

func TestNowWIB(t *testing.T) {
	Convey("Given NowWIB", t, func() {
		result := timeutil.NowWIB()

		Convey("Then the timezone should be WIB", func() {
			So(result.Location().String(), ShouldEqual, "WIB")
		})
	})
}

func TestNowPtr(t *testing.T) {
	Convey("Given NowPtr", t, func() {
		result := timeutil.NowPtr()

		Convey("Then it should return a non-nil pointer", func() {
			So(result, ShouldNotBeNil)
		})

		Convey("Then the time should be recent", func() {
			So(time.Since(*result), ShouldBeLessThan, time.Second)
		})
	})
}
