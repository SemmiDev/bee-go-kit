// Package pagination provides filter, pagination, and paging utilities for
// list/query endpoints.
//
// The Filter struct is designed to be parsed from HTTP query parameters and
// passed through service → repository layers. It supports:
//   - Offset-based pagination (CurrentPage + PerPage)
//   - Keyword search
//   - Sort column + direction
//   - Date-range filtering
//   - Active/inactive filtering
//
// Usage:
//
//	// From a standard net/http request:
//	filter := pagination.FilterFromRequest(r)
//
//	// From any framework that implements InputQuerier:
//	filter := pagination.FilterFromInput(beegoInput)
//
//	// Calculate offset for SQL:
//	offset := filter.GetOffset()  // (page-1) * perPage
//	limit  := filter.GetLimit()
package pagination

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	// AscDirection is the ascending sort direction.
	AscDirection = "asc"

	// DescDirection is the descending sort direction.
	DescDirection = "desc"

	// DefaultPageLimit is the default number of items per page.
	DefaultPageLimit = 20

	// DefaultCurrentPage is the default starting page.
	DefaultCurrentPage = 1

	// DefaultColumnDirection is the default sort direction.
	DefaultColumnDirection = AscDirection

	// UnlimitedPage disables pagination when used as PerPage.
	UnlimitedPage = -1
)

// ---------------------------------------------------------------------------
// Filter
// ---------------------------------------------------------------------------

// Filter represents query parameters for list endpoints. It supports
// pagination, keyword search, sorting, date ranges, and active filtering.
type Filter struct {
	// Pagination
	CurrentPage int `json:"current_page" form:"current_page" query:"current_page" example:"1"`
	PerPage     int `json:"per_page" form:"per_page" query:"per_page" example:"10"`

	// Search
	Keyword string `json:"keyword" form:"keyword" query:"keyword" example:"search"`

	// Sorting
	SortBy        string `json:"sort_by" form:"sort_by" query:"sort_by" example:"name"`
	SortDirection string `json:"sort_direction" form:"sort_direction" query:"sort_direction" example:"asc"`

	// Date Range Filters
	StartDate *time.Time `json:"start_date,omitempty" form:"start_date" query:"start_date" example:"2022-01-01"`
	EndDate   *time.Time `json:"end_date,omitempty" form:"end_date" query:"end_date" example:"2022-12-31"`

	// Active – 1: active only, 0: inactive only, -1: all
	Active int `json:"active" form:"active" query:"active" example:"1"`
}

// NewFilter creates a Filter with sensible defaults.
func NewFilter() Filter {
	return Filter{
		CurrentPage:   DefaultCurrentPage,
		PerPage:       DefaultPageLimit,
		SortDirection: DefaultColumnDirection,
	}
}

// GetLimit returns the number of items per page.
func (f *Filter) GetLimit() int {
	return f.PerPage
}

// GetOffset calculates the SQL offset from the current page and per-page values.
//
// Page 1 → offset 0, Page 2 → offset 10 (if PerPage=10), etc.
func (f *Filter) GetOffset() int {
	return (f.CurrentPage - 1) * f.PerPage
}

// HasKeyword reports whether a search keyword was provided.
func (f *Filter) HasKeyword() bool {
	return f.Keyword != ""
}

// HasSort reports whether a sort column was specified.
func (f *Filter) HasSort() bool {
	return f.SortBy != ""
}

// IsDesc reports whether the sort direction is descending.
func (f *Filter) IsDesc() bool {
	return strings.EqualFold(f.SortDirection, DescDirection)
}

// IsUnlimitedPage reports whether pagination is disabled.
func (f *Filter) IsUnlimitedPage() bool {
	return f.PerPage == UnlimitedPage
}

// HasStartDate reports whether a start date filter was provided.
func (f *Filter) HasStartDate() bool {
	return f.StartDate != nil
}

// HasEndDate reports whether an end date filter was provided.
func (f *Filter) HasEndDate() bool {
	return f.EndDate != nil
}

// HasDateRange reports whether both start and end dates are present.
func (f *Filter) HasDateRange() bool {
	return f.StartDate != nil && f.EndDate != nil
}

// ShowAll reports whether the active filter requests all records.
func (f *Filter) ShowAll() bool {
	return f.Active == -1
}

// Validate normalises Filter fields and replaces invalid values with defaults.
func (f *Filter) Validate() {
	if f.CurrentPage < 1 {
		f.CurrentPage = DefaultCurrentPage
	}
	if f.PerPage < 1 && f.PerPage != UnlimitedPage {
		f.PerPage = DefaultPageLimit
	}
	if f.SortDirection != AscDirection && f.SortDirection != DescDirection {
		f.SortDirection = DefaultColumnDirection
	}
}

// ValidateSortBy checks if the requested sort column is in the allowed list.
// If found, it normalises f.SortBy to the matching column name. Returns true
// when the column is valid or no sort was requested.
func (f *Filter) ValidateSortBy(allowedColumns []string) bool {
	if f.SortBy == "" {
		return true
	}
	for _, col := range allowedColumns {
		if strings.EqualFold(f.SortBy, col) {
			f.SortBy = col
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Parsers – extract Filter from various request sources
// ---------------------------------------------------------------------------

// FilterFromRequest parses filter parameters from a standard net/http request
// using the default parameter names (page, per_page, keyword, etc.).
func FilterFromRequest(r *http.Request) Filter {
	query := r.URL.Query()
	return parseFilter(query.Get)
}

// ---------------------------------------------------------------------------
// Customisable parameter names
// ---------------------------------------------------------------------------

// ParamNames allows clients to customise the query parameter names used when
// parsing filters from HTTP requests.
//
//	params := pagination.DefaultParamNames()
//	params.Page = "pageNo"
//	params.PerPage = "pageSize"
//	filter := pagination.FilterFromRequestWithParams(r, params)
type ParamNames struct {
	Page          string // default: "page"
	PerPage       string // default: "per_page"
	PerPageAlias  string // default: "limit"
	Keyword       string // default: "keyword"
	KeywordAlias1 string // default: "search"
	KeywordAlias2 string // default: "q"
	SortBy        string // default: "sort_by"
	SortByAlias   string // default: "order_by"
	SortDirection string // default: "sort_direction"
	SortDirAlias  string // default: "order"
	StartDate     string // default: "start_date"
	EndDate       string // default: "end_date"
}

// DefaultParamNames returns the default query parameter names.
func DefaultParamNames() ParamNames {
	return ParamNames{
		Page:          "page",
		PerPage:       "per_page",
		PerPageAlias:  "limit",
		Keyword:       "keyword",
		KeywordAlias1: "search",
		KeywordAlias2: "q",
		SortBy:        "sort_by",
		SortByAlias:   "order_by",
		SortDirection: "sort_direction",
		SortDirAlias:  "order",
		StartDate:     "start_date",
		EndDate:       "end_date",
	}
}

// FilterFromRequestWithParams parses filter parameters from a net/http request
// using custom parameter names.
func FilterFromRequestWithParams(r *http.Request, params ParamNames) Filter {
	query := r.URL.Query()
	return parseFilterWithParams(query.Get, params)
}

// InputQuerier is an interface satisfied by any object that can look up a
// query parameter by name (e.g. Beego's Input, Echo's QueryParam, etc.).
type InputQuerier interface {
	Query(string) string
}

// FilterFromInput parses filter parameters from an InputQuerier.
//
// Example with Beego:
//
//	filter := pagination.FilterFromInput(c.Ctx.Input)
func FilterFromInput(input InputQuerier) Filter {
	filter := parseFilter(input.Query)

	// Parse active flag (default: show active only)
	if input.Query("active") != "" {
		if a, err := strconv.Atoi(input.Query("active")); err == nil {
			filter.Active = a
		}
	} else {
		filter.Active = 1
	}

	return filter
}

// parseFilter is the shared parsing logic; getter is a function that returns
// the value for a given query parameter name.
func parseFilter(getter func(string) string) Filter {
	filter := NewFilter()

	// ---- Pagination ----
	if page := getter("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.CurrentPage = p
		}
	}
	if perPage := getter("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil && pp > 0 {
			filter.PerPage = pp
		}
	}
	// "limit" as alias for "per_page"
	if limit := getter("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.PerPage = l
		}
	}

	// ---- Keyword (try "keyword", "search", "q") ----
	filter.Keyword = strings.TrimSpace(getter("keyword"))
	if filter.Keyword == "" {
		filter.Keyword = strings.TrimSpace(getter("search"))
	}
	if filter.Keyword == "" {
		filter.Keyword = strings.TrimSpace(getter("q"))
	}

	// ---- Sorting ----
	filter.SortBy = strings.TrimSpace(getter("sort_by"))
	if filter.SortBy == "" {
		filter.SortBy = strings.TrimSpace(getter("order_by"))
	}
	filter.SortDirection = strings.ToLower(strings.TrimSpace(getter("sort_direction")))
	if filter.SortDirection == "" {
		filter.SortDirection = strings.ToLower(strings.TrimSpace(getter("order")))
	}
	if filter.SortDirection != AscDirection && filter.SortDirection != DescDirection {
		filter.SortDirection = DefaultColumnDirection
	}

	// ---- Date range ----
	if sd := getter("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			filter.StartDate = &t
		}
	}
	if ed := getter("end_date"); ed != "" {
		if t, err := time.Parse("2006-01-02", ed); err == nil {
			filter.EndDate = &t
		}
	}

	return filter
}

// parseFilterWithParams is like parseFilter but uses custom parameter names.
func parseFilterWithParams(getter func(string) string, p ParamNames) Filter {
	filter := NewFilter()

	// ---- Pagination ----
	if page := getter(p.Page); page != "" {
		if v, err := strconv.Atoi(page); err == nil && v > 0 {
			filter.CurrentPage = v
		}
	}
	if perPage := getter(p.PerPage); perPage != "" {
		if v, err := strconv.Atoi(perPage); err == nil && v > 0 {
			filter.PerPage = v
		}
	}
	if p.PerPageAlias != "" {
		if limit := getter(p.PerPageAlias); limit != "" {
			if v, err := strconv.Atoi(limit); err == nil && v > 0 {
				filter.PerPage = v
			}
		}
	}

	// ---- Keyword ----
	filter.Keyword = strings.TrimSpace(getter(p.Keyword))
	if filter.Keyword == "" && p.KeywordAlias1 != "" {
		filter.Keyword = strings.TrimSpace(getter(p.KeywordAlias1))
	}
	if filter.Keyword == "" && p.KeywordAlias2 != "" {
		filter.Keyword = strings.TrimSpace(getter(p.KeywordAlias2))
	}

	// ---- Sorting ----
	filter.SortBy = strings.TrimSpace(getter(p.SortBy))
	if filter.SortBy == "" && p.SortByAlias != "" {
		filter.SortBy = strings.TrimSpace(getter(p.SortByAlias))
	}
	filter.SortDirection = strings.ToLower(strings.TrimSpace(getter(p.SortDirection)))
	if filter.SortDirection == "" && p.SortDirAlias != "" {
		filter.SortDirection = strings.ToLower(strings.TrimSpace(getter(p.SortDirAlias)))
	}
	if filter.SortDirection != AscDirection && filter.SortDirection != DescDirection {
		filter.SortDirection = DefaultColumnDirection
	}

	// ---- Date range ----
	if sd := getter(p.StartDate); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			filter.StartDate = &t
		}
	}
	if ed := getter(p.EndDate); ed != "" {
		if t, err := time.Parse("2006-01-02", ed); err == nil {
			filter.EndDate = &t
		}
	}

	return filter
}
