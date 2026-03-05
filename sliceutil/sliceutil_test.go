package sliceutil_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/sliceutil"
)

// ---------------------------------------------------------------------------
// Transformation
// ---------------------------------------------------------------------------

func TestMap(t *testing.T) {
	Convey("Given Map", t, func() {
		Convey("When applied to a slice of ints", func() {
			result := sliceutil.Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
			So(result, ShouldResemble, []int{2, 4, 6})
		})

		Convey("When applied to convert types", func() {
			result := sliceutil.Map([]int{1, 2}, func(n int) string {
				return "x"
			})
			So(result, ShouldResemble, []string{"x", "x"})
		})

		Convey("When source is nil", func() {
			result := sliceutil.Map[int, int](nil, func(n int) int { return n })
			So(result, ShouldBeNil)
		})

		Convey("When source is empty", func() {
			result := sliceutil.Map([]int{}, func(n int) int { return n })
			So(result, ShouldBeEmpty)
		})
	})
}

func TestFilter(t *testing.T) {
	Convey("Given Filter", t, func() {
		Convey("When filtering even numbers", func() {
			result := sliceutil.Filter([]int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 == 0 })
			So(result, ShouldResemble, []int{2, 4})
		})

		Convey("When no elements match", func() {
			result := sliceutil.Filter([]int{1, 3, 5}, func(n int) bool { return n%2 == 0 })
			So(result, ShouldBeEmpty)
		})

		Convey("When source is nil", func() {
			result := sliceutil.Filter[int](nil, func(n int) bool { return true })
			So(result, ShouldBeNil)
		})
	})
}

func TestReduce(t *testing.T) {
	Convey("Given Reduce", t, func() {
		Convey("When summing numbers", func() {
			result := sliceutil.Reduce([]int{1, 2, 3, 4}, 0, func(acc, n int) int { return acc + n })
			So(result, ShouldEqual, 10)
		})

		Convey("When concatenating strings", func() {
			result := sliceutil.Reduce([]string{"a", "b", "c"}, "", func(acc, s string) string { return acc + s })
			So(result, ShouldEqual, "abc")
		})

		Convey("When source is empty", func() {
			result := sliceutil.Reduce([]int{}, 42, func(acc, n int) int { return acc + n })
			So(result, ShouldEqual, 42)
		})
	})
}

func TestFlatMap(t *testing.T) {
	Convey("Given FlatMap", t, func() {
		Convey("When expanding elements", func() {
			result := sliceutil.FlatMap([]int{1, 2, 3}, func(n int) []int { return []int{n, n * 10} })
			So(result, ShouldResemble, []int{1, 10, 2, 20, 3, 30})
		})

		Convey("When source is empty", func() {
			result := sliceutil.FlatMap([]int{}, func(n int) []int { return []int{n} })
			So(result, ShouldBeNil)
		})
	})
}

// ---------------------------------------------------------------------------
// Lookup
// ---------------------------------------------------------------------------

func TestContains(t *testing.T) {
	Convey("Given Contains", t, func() {
		Convey("When element exists", func() {
			So(sliceutil.Contains([]string{"a", "b", "c"}, "b"), ShouldBeTrue)
		})

		Convey("When element does not exist", func() {
			So(sliceutil.Contains([]string{"a", "b"}, "z"), ShouldBeFalse)
		})

		Convey("When slice is empty", func() {
			So(sliceutil.Contains([]int{}, 1), ShouldBeFalse)
		})
	})
}

func TestIndexOf(t *testing.T) {
	Convey("Given IndexOf", t, func() {
		Convey("When element exists", func() {
			So(sliceutil.IndexOf([]string{"a", "b", "c"}, "b"), ShouldEqual, 1)
		})

		Convey("When element does not exist", func() {
			So(sliceutil.IndexOf([]string{"a", "b"}, "z"), ShouldEqual, -1)
		})
	})
}

func TestFind(t *testing.T) {
	Convey("Given Find", t, func() {
		Convey("When element matching predicate exists", func() {
			val, found := sliceutil.Find([]int{1, 2, 3, 4}, func(n int) bool { return n > 2 })
			So(found, ShouldBeTrue)
			So(val, ShouldEqual, 3)
		})

		Convey("When no element matches", func() {
			_, found := sliceutil.Find([]int{1, 2}, func(n int) bool { return n > 10 })
			So(found, ShouldBeFalse)
		})
	})
}

// ---------------------------------------------------------------------------
// Grouping & Partitioning
// ---------------------------------------------------------------------------

func TestGroupBy(t *testing.T) {
	Convey("Given GroupBy", t, func() {
		Convey("When grouping by a key", func() {
			type Item struct {
				Category string
				Name     string
			}
			items := []Item{
				{"fruit", "apple"},
				{"veg", "carrot"},
				{"fruit", "banana"},
			}
			result := sliceutil.GroupBy(items, func(i Item) string { return i.Category })
			So(result, ShouldHaveLength, 2)
			So(result["fruit"], ShouldHaveLength, 2)
			So(result["veg"], ShouldHaveLength, 1)
		})
	})
}

func TestChunk(t *testing.T) {
	Convey("Given Chunk", t, func() {
		Convey("When size divides evenly", func() {
			result := sliceutil.Chunk([]int{1, 2, 3, 4}, 2)
			So(result, ShouldResemble, [][]int{{1, 2}, {3, 4}})
		})

		Convey("When last chunk is smaller", func() {
			result := sliceutil.Chunk([]int{1, 2, 3, 4, 5}, 2)
			So(result, ShouldHaveLength, 3)
			So(result[2], ShouldResemble, []int{5})
		})

		Convey("When size is 0 or negative", func() {
			So(sliceutil.Chunk([]int{1, 2}, 0), ShouldBeNil)
			So(sliceutil.Chunk([]int{1, 2}, -1), ShouldBeNil)
		})

		Convey("When source is empty", func() {
			So(sliceutil.Chunk([]int{}, 5), ShouldBeNil)
		})
	})
}

// ---------------------------------------------------------------------------
// Deduplication & Set operations
// ---------------------------------------------------------------------------

func TestUnique(t *testing.T) {
	Convey("Given Unique", t, func() {
		Convey("When duplicates exist", func() {
			result := sliceutil.Unique([]int{1, 2, 2, 3, 1})
			So(result, ShouldResemble, []int{1, 2, 3})
		})

		Convey("When no duplicates", func() {
			result := sliceutil.Unique([]int{1, 2, 3})
			So(result, ShouldResemble, []int{1, 2, 3})
		})

		Convey("When source is nil", func() {
			So(sliceutil.Unique[int](nil), ShouldBeNil)
		})
	})
}

func TestDifference(t *testing.T) {
	Convey("Given Difference", t, func() {
		Convey("When sets overlap", func() {
			result := sliceutil.Difference([]int{1, 2, 3, 4}, []int{2, 4})
			So(result, ShouldResemble, []int{1, 3})
		})

		Convey("When sets are identical", func() {
			result := sliceutil.Difference([]int{1, 2}, []int{1, 2})
			So(result, ShouldBeNil)
		})
	})
}

func TestIntersect(t *testing.T) {
	Convey("Given Intersect", t, func() {
		Convey("When sets overlap", func() {
			result := sliceutil.Intersect([]int{1, 2, 3}, []int{2, 3, 4})
			So(result, ShouldResemble, []int{2, 3})
		})

		Convey("When sets have no overlap", func() {
			result := sliceutil.Intersect([]int{1, 2}, []int{3, 4})
			So(result, ShouldBeNil)
		})
	})
}

// ---------------------------------------------------------------------------
// Convenience
// ---------------------------------------------------------------------------

func TestToMap(t *testing.T) {
	Convey("Given ToMap", t, func() {
		type User struct {
			ID   string
			Name string
		}
		users := []User{{ID: "1", Name: "Alice"}, {ID: "2", Name: "Bob"}}
		result := sliceutil.ToMap(users, func(u User) string { return u.ID })

		Convey("Then it should create a map keyed by ID", func() {
			So(result, ShouldHaveLength, 2)
			So(result["1"].Name, ShouldEqual, "Alice")
			So(result["2"].Name, ShouldEqual, "Bob")
		})
	})
}

func TestKeys(t *testing.T) {
	Convey("Given Keys", t, func() {
		m := map[string]int{"a": 1, "b": 2}
		result := sliceutil.Keys(m)

		Convey("Then it should extract all keys", func() {
			So(result, ShouldHaveLength, 2)
			So(result, ShouldContain, "a")
			So(result, ShouldContain, "b")
		})
	})
}

func TestValues(t *testing.T) {
	Convey("Given Values", t, func() {
		m := map[string]int{"a": 1, "b": 2}
		result := sliceutil.Values(m)

		Convey("Then it should extract all values", func() {
			So(result, ShouldHaveLength, 2)
			So(result, ShouldContain, 1)
			So(result, ShouldContain, 2)
		})
	})
}

func TestIsEmpty(t *testing.T) {
	Convey("Given IsEmpty", t, func() {
		So(sliceutil.IsEmpty[int](nil), ShouldBeTrue)
		So(sliceutil.IsEmpty([]int{}), ShouldBeTrue)
		So(sliceutil.IsEmpty([]int{1}), ShouldBeFalse)
	})
}
