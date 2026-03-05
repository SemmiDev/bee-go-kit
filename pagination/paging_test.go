package pagination_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/pagination"
)

func TestNewPaging(t *testing.T) {
	Convey("Given NewPaging", t, func() {
		Convey("When on the first page with 100 total records", func() {
			p, err := pagination.NewPaging(1, 10, 100)

			Convey("Then it should return valid paging metadata", func() {
				So(err, ShouldBeNil)
				So(p.CurrentPage, ShouldEqual, 1)
				So(p.PerPage, ShouldEqual, 10)
				So(p.TotalData, ShouldEqual, 100)
				So(p.LastPage, ShouldEqual, 10)
				So(p.From, ShouldEqual, 1)
				So(p.To, ShouldEqual, 10)
				So(p.TotalDataInCurrentPage, ShouldEqual, 10)
				So(p.HasPreviousPage, ShouldBeFalse)
				So(p.HasNextPage, ShouldBeTrue)
			})
		})

		Convey("When on a middle page", func() {
			p, err := pagination.NewPaging(5, 10, 100)

			Convey("Then it should have both previous and next pages", func() {
				So(err, ShouldBeNil)
				So(p.HasPreviousPage, ShouldBeTrue)
				So(p.HasNextPage, ShouldBeTrue)
				So(p.From, ShouldEqual, 41)
				So(p.To, ShouldEqual, 50)
			})
		})

		Convey("When on the last page", func() {
			p, err := pagination.NewPaging(10, 10, 100)

			Convey("Then HasNextPage should be false", func() {
				So(err, ShouldBeNil)
				So(p.HasNextPage, ShouldBeFalse)
				So(p.HasPreviousPage, ShouldBeTrue)
			})
		})

		Convey("When totalData is 0", func() {
			p, err := pagination.NewPaging(1, 10, 0)

			Convey("Then it should return an empty first page", func() {
				So(err, ShouldBeNil)
				So(p.TotalData, ShouldEqual, 0)
				So(p.From, ShouldEqual, 0)
				So(p.To, ShouldEqual, 0)
				So(p.TotalDataInCurrentPage, ShouldEqual, 0)
				So(p.HasPreviousPage, ShouldBeFalse)
				So(p.HasNextPage, ShouldBeFalse)
			})
		})

		Convey("When perPage is UnlimitedPage (-1)", func() {
			p, err := pagination.NewPaging(1, -1, 50)

			Convey("Then all data should be on a single page", func() {
				So(err, ShouldBeNil)
				So(p.LastPage, ShouldEqual, 1)
				So(p.TotalDataInCurrentPage, ShouldEqual, 50)
				So(p.From, ShouldEqual, 1)
				So(p.To, ShouldEqual, 50)
			})
		})

		Convey("When perPage is 0 (invalid)", func() {
			_, err := pagination.NewPaging(1, 0, 50)

			Convey("Then it should return ErrPaging", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, pagination.ErrPaging)
			})
		})

		Convey("When currentPage exceeds lastPage", func() {
			p, err := pagination.NewPaging(999, 10, 25)

			Convey("Then current page should be clamped to last page", func() {
				So(err, ShouldBeNil)
				So(p.CurrentPage, ShouldEqual, 3) // 25/10 = 3 pages
			})
		})

		Convey("When totalData is not evenly divisible by perPage", func() {
			p, err := pagination.NewPaging(1, 10, 25)

			Convey("Then lastPage should round up", func() {
				So(err, ShouldBeNil)
				So(p.LastPage, ShouldEqual, 3) // ceil(25/10) = 3
			})
		})
	})
}
