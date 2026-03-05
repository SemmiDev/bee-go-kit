package pagination

import "errors"

// ---------------------------------------------------------------------------
// Paging
// ---------------------------------------------------------------------------

// ErrPaging is returned when PerPage is <= 0 or offset is negative.
var ErrPaging = errors.New("per_page must be > 0 and offset must be >= 0")

// Paging contains computed pagination metadata that should be included in list
// responses so the client can navigate through pages.
type Paging struct {
	HasPreviousPage        bool `json:"has_previous_page" example:"true"`
	HasNextPage            bool `json:"has_next_page" example:"true"`
	CurrentPage            int  `json:"current_page" example:"1"`
	PerPage                int  `json:"per_page" example:"10"`
	TotalData              int  `json:"total_data" example:"100"`
	TotalDataInCurrentPage int  `json:"total_data_in_current_page" example:"10"`
	LastPage               int  `json:"last_page" example:"10"`
	From                   int  `json:"from" example:"1"`
	To                     int  `json:"to" example:"10"`
}

// NewPaging calculates paging metadata from the current page, per-page limit,
// and total number of records.
//
// If perPage is UnlimitedPage (-1), all records are considered to be on a
// single page. Returns ErrPaging for invalid inputs.
func NewPaging(currentPage, perPage, totalData int) (*Paging, error) {
	// Unlimited page mode – everything on one page.
	if perPage == UnlimitedPage {
		return &Paging{
			CurrentPage:            currentPage,
			PerPage:                perPage,
			TotalData:              totalData,
			LastPage:               1,
			From:                   1,
			To:                     totalData,
			TotalDataInCurrentPage: totalData,
		}, nil
	}

	// No data – return an empty first page.
	if totalData == 0 {
		return &Paging{
			HasPreviousPage:        false,
			HasNextPage:            false,
			CurrentPage:            1,
			PerPage:                perPage,
			TotalData:              0,
			TotalDataInCurrentPage: 0,
			LastPage:               1,
			From:                   0,
			To:                     0,
		}, nil
	}

	offset := (currentPage - 1) * perPage

	if perPage <= 0 || offset < 0 {
		return nil, ErrPaging
	}

	// Calculate last page (ceiling division).
	lastPage := totalData / perPage
	if totalData%perPage != 0 {
		lastPage++
	}

	to := min(offset+perPage, totalData)
	from := 0
	if to > offset {
		from = offset + 1
	}

	// Clamp current page to last page.
	if currentPage > lastPage {
		currentPage = lastPage
	}

	totalDataInCurrentPage := to - offset

	return &Paging{
		HasPreviousPage:        currentPage > 1,
		HasNextPage:            currentPage < lastPage,
		CurrentPage:            currentPage,
		PerPage:                perPage,
		TotalData:              totalData,
		LastPage:               lastPage,
		From:                   from,
		To:                     to,
		TotalDataInCurrentPage: totalDataInCurrentPage,
	}, nil
}
