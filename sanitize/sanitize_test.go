package sanitize_test

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/sanitize"
)

// ---------------------------------------------------------------------------
// Composable pipeline (extensibility)
// ---------------------------------------------------------------------------

func TestChain(t *testing.T) {
	Convey("Given Chain", t, func() {
		Convey("When composing multiple sanitisers", func() {
			pipeline := sanitize.Chain(sanitize.StringNoEscape, strings.ToLower)
			result := pipeline("  HELLO World  ")

			Convey("Then functions should be applied left to right", func() {
				So(result, ShouldEqual, "hello world")
			})
		})

		Convey("When chaining with custom functions", func() {
			addPrefix := func(s string) string { return "PREFIX_" + s }
			pipeline := sanitize.Chain(sanitize.StringNoEscape, addPrefix)
			So(pipeline("  test  "), ShouldEqual, "PREFIX_test")
		})

		Convey("When chain is empty", func() {
			pipeline := sanitize.Chain()
			So(pipeline("unchanged"), ShouldEqual, "unchanged")
		})
	})
}

func TestApply(t *testing.T) {
	Convey("Given Apply", t, func() {
		Convey("When applying multiple functions", func() {
			result := sanitize.Apply("  HELLO  ", sanitize.StringNoEscape, strings.ToLower)
			So(result, ShouldEqual, "hello")
		})

		Convey("When no functions are provided", func() {
			So(sanitize.Apply("unchanged"), ShouldEqual, "unchanged")
		})
	})
}

// ---------------------------------------------------------------------------
// Built-in sanitisers
// ---------------------------------------------------------------------------

func TestString(t *testing.T) {
	Convey("Given the String sanitiser", t, func() {
		Convey("When input has leading/trailing whitespace", func() {
			Convey("Then it should trim them", func() {
				So(sanitize.String("  hello  "), ShouldEqual, "hello")
			})
		})

		Convey("When input contains HTML special characters", func() {
			Convey("Then it should escape them", func() {
				So(sanitize.String("<script>alert('xss')</script>"), ShouldNotContainSubstring, "<script>")
				So(sanitize.String("a & b"), ShouldContainSubstring, "&amp;")
			})
		})

		Convey("When input is empty", func() {
			So(sanitize.String(""), ShouldBeEmpty)
		})
	})
}

func TestStringNoEscape(t *testing.T) {
	Convey("Given the StringNoEscape sanitiser", t, func() {
		Convey("When input has whitespace", func() {
			So(sanitize.StringNoEscape("  hello  "), ShouldEqual, "hello")
		})

		Convey("When input contains HTML characters", func() {
			Convey("Then it should preserve them", func() {
				So(sanitize.StringNoEscape("<b>bold</b>"), ShouldEqual, "<b>bold</b>")
			})
		})
	})
}

func TestForSQLLike(t *testing.T) {
	Convey("Given the ForSQLLike sanitiser", t, func() {
		Convey("When input contains % wildcard", func() {
			So(sanitize.ForSQLLike("100%"), ShouldEqual, "100[%]")
		})

		Convey("When input contains _ wildcard", func() {
			So(sanitize.ForSQLLike("a_b"), ShouldEqual, "a[_]b")
		})

		Convey("When input contains [ bracket", func() {
			So(sanitize.ForSQLLike("a[b"), ShouldEqual, "a[[]b")
		})

		Convey("When input has leading whitespace", func() {
			So(sanitize.ForSQLLike("  test  "), ShouldEqual, "test")
		})

		Convey("When input is safe", func() {
			So(sanitize.ForSQLLike("hello"), ShouldEqual, "hello")
		})
	})
}

func TestEmail(t *testing.T) {
	Convey("Given the Email sanitiser", t, func() {
		Convey("When email is valid", func() {
			So(sanitize.Email("User@Example.COM"), ShouldEqual, "user@example.com")
		})

		Convey("When email has whitespace", func() {
			So(sanitize.Email("  user@example.com  "), ShouldEqual, "user@example.com")
		})

		Convey("When email is invalid", func() {
			So(sanitize.Email("not-an-email"), ShouldBeEmpty)
			So(sanitize.Email("@missing.local"), ShouldBeEmpty)
			So(sanitize.Email(""), ShouldBeEmpty)
		})
	})
}

func TestAlphanumeric(t *testing.T) {
	Convey("Given the Alphanumeric sanitiser", t, func() {
		Convey("When input has special characters", func() {
			So(sanitize.Alphanumeric("abc-123!@#"), ShouldEqual, "abc123")
		})

		Convey("When input is already clean", func() {
			So(sanitize.Alphanumeric("hello123"), ShouldEqual, "hello123")
		})

		Convey("When input is empty", func() {
			So(sanitize.Alphanumeric(""), ShouldBeEmpty)
		})
	})
}

func TestNumeric(t *testing.T) {
	Convey("Given the Numeric sanitiser", t, func() {
		Convey("When input has non-digit characters", func() {
			So(sanitize.Numeric("+62-812-3456"), ShouldEqual, "628123456")
		})

		Convey("When input is all digits", func() {
			So(sanitize.Numeric("12345"), ShouldEqual, "12345")
		})
	})
}

func TestFilename(t *testing.T) {
	Convey("Given the Filename sanitiser", t, func() {
		Convey("When filename contains path traversal", func() {
			So(sanitize.Filename("../../etc/passwd"), ShouldNotContainSubstring, "..")
			So(sanitize.Filename("../../etc/passwd"), ShouldNotContainSubstring, "/")
		})

		Convey("When filename contains null bytes", func() {
			So(sanitize.Filename("file\x00name.txt"), ShouldEqual, "filename.txt")
		})

		Convey("When filename contains dangerous characters", func() {
			result := sanitize.Filename("file<>:\"|?*.txt")
			So(result, ShouldEqual, "file.txt")
		})

		Convey("When filename is safe", func() {
			So(sanitize.Filename("report.pdf"), ShouldEqual, "report.pdf")
		})
	})
}

func TestPtr(t *testing.T) {
	Convey("Given the Ptr sanitiser", t, func() {
		Convey("When pointer is nil", func() {
			So(sanitize.Ptr(nil), ShouldBeNil)
		})

		Convey("When pointed string is empty after trim", func() {
			s := "   "
			So(sanitize.Ptr(&s), ShouldBeNil)
		})

		Convey("When pointed string has whitespace", func() {
			s := "  hello  "
			result := sanitize.Ptr(&s)
			So(result, ShouldNotBeNil)
			So(*result, ShouldEqual, "hello")
		})
	})
}
