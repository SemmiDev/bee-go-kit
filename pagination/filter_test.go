package pagination_test

import (
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/pagination"
)

// ---------------------------------------------------------------------------
// Filter defaults & getters
// ---------------------------------------------------------------------------

func TestNewFilter(t *testing.T) {
	Convey("Given NewFilter", t, func() {
		f := pagination.NewFilter()

		Convey("Then it should have sensible defaults", func() {
			So(f.CurrentPage, ShouldEqual, 1)
			So(f.PerPage, ShouldEqual, 20)
			So(f.SortDirection, ShouldEqual, "asc")
		})
	})
}

func TestFilter_GetLimit(t *testing.T) {
	Convey("Given a Filter", t, func() {
		f := pagination.NewFilter()
		So(f.GetLimit(), ShouldEqual, 20)
	})
}

func TestFilter_GetOffset(t *testing.T) {
	Convey("Given GetOffset", t, func() {
		Convey("When page is 1", func() {
			f := pagination.Filter{CurrentPage: 1, PerPage: 10}
			So(f.GetOffset(), ShouldEqual, 0)
		})

		Convey("When page is 3 with 10 per page", func() {
			f := pagination.Filter{CurrentPage: 3, PerPage: 10}
			So(f.GetOffset(), ShouldEqual, 20)
		})
	})
}

func TestFilter_HasKeyword(t *testing.T) {
	Convey("Given HasKeyword", t, func() {
		Convey("When keyword is set", func() {
			f := pagination.Filter{Keyword: "search"}
			So(f.HasKeyword(), ShouldBeTrue)
		})

		Convey("When keyword is empty", func() {
			f := pagination.Filter{}
			So(f.HasKeyword(), ShouldBeFalse)
		})
	})
}

func TestFilter_HasSort(t *testing.T) {
	Convey("Given HasSort", t, func() {
		f := pagination.Filter{SortBy: "name"}
		So(f.HasSort(), ShouldBeTrue)

		f2 := pagination.Filter{}
		So(f2.HasSort(), ShouldBeFalse)
	})
}

func TestFilter_IsDesc(t *testing.T) {
	Convey("Given IsDesc", t, func() {
		Convey("When direction is desc", func() {
			f := pagination.Filter{SortDirection: "desc"}
			So(f.IsDesc(), ShouldBeTrue)
		})

		Convey("When direction is asc", func() {
			f := pagination.Filter{SortDirection: "asc"}
			So(f.IsDesc(), ShouldBeFalse)
		})
	})
}

func TestFilter_IsUnlimitedPage(t *testing.T) {
	Convey("Given IsUnlimitedPage", t, func() {
		f := pagination.Filter{PerPage: -1}
		So(f.IsUnlimitedPage(), ShouldBeTrue)

		f2 := pagination.Filter{PerPage: 10}
		So(f2.IsUnlimitedPage(), ShouldBeFalse)
	})
}

func TestFilter_ShowAll(t *testing.T) {
	Convey("Given ShowAll", t, func() {
		f := pagination.Filter{Active: -1}
		So(f.ShowAll(), ShouldBeTrue)

		f2 := pagination.Filter{Active: 1}
		So(f2.ShowAll(), ShouldBeFalse)
	})
}

// ---------------------------------------------------------------------------
// Validate
// ---------------------------------------------------------------------------

func TestFilter_Validate(t *testing.T) {
	Convey("Given Validate", t, func() {
		Convey("When page is 0", func() {
			f := pagination.Filter{CurrentPage: 0, PerPage: 10, SortDirection: "asc"}
			f.Validate()
			So(f.CurrentPage, ShouldEqual, 1)
		})

		Convey("When perPage is 0", func() {
			f := pagination.Filter{CurrentPage: 1, PerPage: 0, SortDirection: "asc"}
			f.Validate()
			So(f.PerPage, ShouldEqual, 20)
		})

		Convey("When perPage is UnlimitedPage", func() {
			f := pagination.Filter{CurrentPage: 1, PerPage: -1, SortDirection: "asc"}
			f.Validate()
			So(f.PerPage, ShouldEqual, -1) // should stay
		})

		Convey("When direction is invalid", func() {
			f := pagination.Filter{CurrentPage: 1, PerPage: 10, SortDirection: "sideways"}
			f.Validate()
			So(f.SortDirection, ShouldEqual, "asc")
		})
	})
}

// ---------------------------------------------------------------------------
// ValidateSortBy
// ---------------------------------------------------------------------------

func TestFilter_ValidateSortBy(t *testing.T) {
	Convey("Given ValidateSortBy", t, func() {
		Convey("When sort column is in allowed list", func() {
			f := pagination.Filter{SortBy: "NAME"}
			result := f.ValidateSortBy([]string{"name", "age"})
			So(result, ShouldBeTrue)
			So(f.SortBy, ShouldEqual, "name") // normalised
		})

		Convey("When sort column is not in allowed list", func() {
			f := pagination.Filter{SortBy: "evil"}
			result := f.ValidateSortBy([]string{"name"})
			So(result, ShouldBeFalse)
		})

		Convey("When SortBy is empty", func() {
			f := pagination.Filter{}
			result := f.ValidateSortBy([]string{"name"})
			So(result, ShouldBeTrue)
		})
	})
}

// ---------------------------------------------------------------------------
// FilterFromRequest
// ---------------------------------------------------------------------------

func TestFilterFromRequest(t *testing.T) {
	Convey("Given FilterFromRequest", t, func() {
		Convey("When all query parameters are present", func() {
			u, _ := url.Parse("http://example.com/api?page=3&per_page=15&keyword=test&sort_by=name&sort_direction=desc&start_date=2026-01-01&end_date=2026-12-31")
			r := &http.Request{URL: u}
			f := pagination.FilterFromRequest(r)

			So(f.CurrentPage, ShouldEqual, 3)
			So(f.PerPage, ShouldEqual, 15)
			So(f.Keyword, ShouldEqual, "test")
			So(f.SortBy, ShouldEqual, "name")
			So(f.SortDirection, ShouldEqual, "desc")
			So(f.HasStartDate(), ShouldBeTrue)
			So(f.HasEndDate(), ShouldBeTrue)
			So(f.HasDateRange(), ShouldBeTrue)
		})

		Convey("When no query parameters are present", func() {
			u, _ := url.Parse("http://example.com/api")
			r := &http.Request{URL: u}
			f := pagination.FilterFromRequest(r)

			So(f.CurrentPage, ShouldEqual, 1)
			So(f.PerPage, ShouldEqual, 20)
			So(f.SortDirection, ShouldEqual, "asc")
			So(f.HasKeyword(), ShouldBeFalse)
		})

		Convey("When keyword aliases are used", func() {
			u, _ := url.Parse("http://example.com/api?q=hello")
			r := &http.Request{URL: u}
			f := pagination.FilterFromRequest(r)
			So(f.Keyword, ShouldEqual, "hello")
		})
	})
}

// ---------------------------------------------------------------------------
// FilterFromRequestWithParams (extensibility)
// ---------------------------------------------------------------------------

func TestFilterFromRequestWithParams(t *testing.T) {
	Convey("Given FilterFromRequestWithParams", t, func() {
		Convey("When custom param names are used", func() {
			params := pagination.ParamNames{
				Page:          "pageNo",
				PerPage:       "pageSize",
				PerPageAlias:  "",
				Keyword:       "search",
				KeywordAlias1: "",
				KeywordAlias2: "",
				SortBy:        "orderBy",
				SortByAlias:   "",
				SortDirection: "orderDir",
				SortDirAlias:  "",
				StartDate:     "from",
				EndDate:       "to",
			}

			u, _ := url.Parse("http://example.com/api?pageNo=5&pageSize=30&search=hello&orderBy=name&orderDir=desc&from=2026-01-01&to=2026-12-31")
			r := &http.Request{URL: u}
			f := pagination.FilterFromRequestWithParams(r, params)

			So(f.CurrentPage, ShouldEqual, 5)
			So(f.PerPage, ShouldEqual, 30)
			So(f.Keyword, ShouldEqual, "hello")
			So(f.SortBy, ShouldEqual, "name")
			So(f.SortDirection, ShouldEqual, "desc")
			So(f.HasDateRange(), ShouldBeTrue)
		})

		Convey("When using DefaultParamNames", func() {
			params := pagination.DefaultParamNames()

			u, _ := url.Parse("http://example.com/api?page=2&per_page=10")
			r := &http.Request{URL: u}
			f := pagination.FilterFromRequestWithParams(r, params)

			Convey("Then it should behave like FilterFromRequest", func() {
				So(f.CurrentPage, ShouldEqual, 2)
				So(f.PerPage, ShouldEqual, 10)
			})
		})
	})
}
