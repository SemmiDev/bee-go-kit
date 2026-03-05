package stringutil_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/stringutil"
)

// ---------------------------------------------------------------------------
// Truncation
// ---------------------------------------------------------------------------

func TestTruncate(t *testing.T) {
	Convey("Given Truncate", t, func() {
		Convey("When string fits within maxLen", func() {
			So(stringutil.Truncate("hello", 10), ShouldEqual, "hello")
		})

		Convey("When string exceeds maxLen", func() {
			So(stringutil.Truncate("Hello World", 5), ShouldEqual, "Hello...")
		})

		Convey("When maxLen equals string length", func() {
			So(stringutil.Truncate("abc", 3), ShouldEqual, "abc")
		})

		Convey("When string contains Unicode", func() {
			So(stringutil.Truncate("héllo wörld", 5), ShouldEqual, "héllo...")
		})
	})
}

func TestTruncateWithSuffix(t *testing.T) {
	Convey("Given TruncateWithSuffix", t, func() {
		Convey("When string exceeds maxLen", func() {
			So(stringutil.TruncateWithSuffix("Hello World", 5, "→"), ShouldEqual, "Hello→")
		})

		Convey("When string fits", func() {
			So(stringutil.TruncateWithSuffix("Hi", 5, "→"), ShouldEqual, "Hi")
		})
	})
}

// ---------------------------------------------------------------------------
// Masking
// ---------------------------------------------------------------------------

func TestMaskEmail(t *testing.T) {
	Convey("Given MaskEmail", t, func() {
		Convey("When local part has more than 2 chars", func() {
			result := stringutil.MaskEmail("john.doe@example.com")
			So(result, ShouldStartWith, "j")
			So(result, ShouldEndWith, "e@example.com")
			So(result, ShouldContainSubstring, "***")
		})

		Convey("When local part has 2 chars", func() {
			result := stringutil.MaskEmail("ab@x.com")
			So(result, ShouldEqual, "a*@x.com")
		})

		Convey("When input has no @", func() {
			So(stringutil.MaskEmail("noemail"), ShouldEqual, "noemail")
		})
	})
}

func TestMaskPhone(t *testing.T) {
	Convey("Given MaskPhone", t, func() {
		Convey("When phone has more than 4 chars", func() {
			result := stringutil.MaskPhone("+6281234567890")
			So(result, ShouldEndWith, "7890")
			So(result, ShouldStartWith, "***")
		})

		Convey("When phone has 4 or fewer chars", func() {
			So(stringutil.MaskPhone("1234"), ShouldEqual, "1234")
		})
	})
}

// ---------------------------------------------------------------------------
// Slug
// ---------------------------------------------------------------------------

func TestSlugify(t *testing.T) {
	Convey("Given Slugify", t, func() {
		Convey("When input has special characters", func() {
			So(stringutil.Slugify("Hello World!"), ShouldEqual, "hello-world")
		})

		Convey("When input has leading/trailing spaces", func() {
			So(stringutil.Slugify("  Foo BAR  "), ShouldEqual, "foo-bar")
		})

		Convey("When input is already a slug", func() {
			So(stringutil.Slugify("already-a-slug"), ShouldEqual, "already-a-slug")
		})
	})
}

// ---------------------------------------------------------------------------
// Case conversion
// ---------------------------------------------------------------------------

func TestToSnakeCase(t *testing.T) {
	Convey("Given ToSnakeCase", t, func() {
		Convey("When input is CamelCase", func() {
			So(stringutil.ToSnakeCase("MyFieldName"), ShouldEqual, "my_field_name")
		})

		Convey("When input has acronyms", func() {
			So(stringutil.ToSnakeCase("HTTPStatusCode"), ShouldEqual, "http_status_code")
		})

		Convey("When input is already lowercase", func() {
			So(stringutil.ToSnakeCase("lowercase"), ShouldEqual, "lowercase")
		})

		Convey("When input has digits", func() {
			So(stringutil.ToSnakeCase("Go118Feature"), ShouldEqual, "go118_feature")
		})
	})
}

// ---------------------------------------------------------------------------
// Misc
// ---------------------------------------------------------------------------

func TestCoalesce(t *testing.T) {
	Convey("Given Coalesce", t, func() {
		Convey("When the first value is non-empty", func() {
			So(stringutil.Coalesce("first", "second"), ShouldEqual, "first")
		})

		Convey("When earlier values are empty", func() {
			So(stringutil.Coalesce("", "", "third"), ShouldEqual, "third")
		})

		Convey("When all values are empty", func() {
			So(stringutil.Coalesce("", ""), ShouldBeEmpty)
		})
	})
}

func TestContainsAny(t *testing.T) {
	Convey("Given ContainsAny", t, func() {
		Convey("When string contains one of the substrings", func() {
			So(stringutil.ContainsAny("Hello World", "world", "foo"), ShouldBeTrue)
		})

		Convey("When string contains none of the substrings", func() {
			So(stringutil.ContainsAny("Hello World", "xyz", "abc"), ShouldBeFalse)
		})
	})
}

func TestDefaultIfEmpty(t *testing.T) {
	Convey("Given DefaultIfEmpty", t, func() {
		Convey("When string is empty", func() {
			So(stringutil.DefaultIfEmpty("", "fallback"), ShouldEqual, "fallback")
		})

		Convey("When string is non-empty", func() {
			So(stringutil.DefaultIfEmpty("present", "fallback"), ShouldEqual, "present")
		})
	})
}

func TestReverse(t *testing.T) {
	Convey("Given Reverse", t, func() {
		Convey("When input is ASCII", func() {
			So(stringutil.Reverse("hello"), ShouldEqual, "olleh")
		})

		Convey("When input is empty", func() {
			So(stringutil.Reverse(""), ShouldBeEmpty)
		})

		Convey("When input has Unicode", func() {
			So(stringutil.Reverse("héllo"), ShouldEqual, "olléh")
		})
	})
}
