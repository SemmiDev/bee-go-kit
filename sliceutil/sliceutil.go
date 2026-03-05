// Package sliceutil provides generic utility functions for working with
// slices. All functions use Go 1.18+ generics.
//
// Usage:
//
//	ids := sliceutil.Map(users, func(u User) string { return u.ID })
//	active := sliceutil.Filter(users, func(u User) bool { return u.Active })
//	if sliceutil.Contains(roles, "admin") { ... }
//	chunks := sliceutil.Chunk(items, 100) // batch processing
package sliceutil

// ---------------------------------------------------------------------------
// Transformation
// ---------------------------------------------------------------------------

// Map applies fn to each element of src and returns a new slice of results.
//
//	names := sliceutil.Map(users, func(u User) string { return u.Name })
func Map[T any, U any](src []T, fn func(T) U) []U {
	if src == nil {
		return nil
	}
	result := make([]U, len(src))
	for i, v := range src {
		result[i] = fn(v)
	}
	return result
}

// Filter returns a new slice containing only elements for which fn returns true.
//
//	admins := sliceutil.Filter(users, func(u User) bool { return u.Role == "admin" })
func Filter[T any](src []T, fn func(T) bool) []T {
	if src == nil {
		return nil
	}
	result := make([]T, 0, len(src)/2)
	for _, v := range src {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce reduces a slice to a single value by applying fn to an accumulator
// and each element.
//
//	total := sliceutil.Reduce(prices, 0, func(acc int, p Price) int { return acc + p.Amount })
func Reduce[T any, U any](src []T, initial U, fn func(U, T) U) U {
	acc := initial
	for _, v := range src {
		acc = fn(acc, v)
	}
	return acc
}

// FlatMap applies fn to each element and flattens the results into a single slice.
func FlatMap[T any, U any](src []T, fn func(T) []U) []U {
	var result []U
	for _, v := range src {
		result = append(result, fn(v)...)
	}
	return result
}

// ---------------------------------------------------------------------------
// Lookup
// ---------------------------------------------------------------------------

// Contains reports whether v is present in the slice.
//
//	if sliceutil.Contains(roles, "admin") { ... }
func Contains[T comparable](src []T, v T) bool {
	for _, item := range src {
		if item == v {
			return true
		}
	}
	return false
}

// IndexOf returns the first index of v in src, or -1 if not found.
func IndexOf[T comparable](src []T, v T) int {
	for i, item := range src {
		if item == v {
			return i
		}
	}
	return -1
}

// Find returns the first element for which fn returns true, along with a bool
// indicating whether an element was found.
func Find[T any](src []T, fn func(T) bool) (T, bool) {
	for _, v := range src {
		if fn(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// ---------------------------------------------------------------------------
// Grouping & Partitioning
// ---------------------------------------------------------------------------

// GroupBy groups elements by a key extracted via fn.
//
//	byDept := sliceutil.GroupBy(employees, func(e Employee) string { return e.Department })
func GroupBy[T any, K comparable](src []T, fn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range src {
		key := fn(v)
		result[key] = append(result[key], v)
	}
	return result
}

// Chunk splits a slice into chunks of the given size. The last chunk may be
// smaller than size. Useful for batch processing.
//
//	for _, batch := range sliceutil.Chunk(items, 100) { processBatch(batch) }
func Chunk[T any](src []T, size int) [][]T {
	if size <= 0 || len(src) == 0 {
		return nil
	}
	chunks := make([][]T, 0, (len(src)+size-1)/size)
	for i := 0; i < len(src); i += size {
		end := i + size
		if end > len(src) {
			end = len(src)
		}
		chunks = append(chunks, src[i:end])
	}
	return chunks
}

// ---------------------------------------------------------------------------
// Deduplication & Set operations
// ---------------------------------------------------------------------------

// Unique returns a new slice with duplicate elements removed, preserving
// the order of first occurrence.
//
//	uniq := sliceutil.Unique([]int{1, 2, 2, 3, 1}) // [1, 2, 3]
func Unique[T comparable](src []T) []T {
	if src == nil {
		return nil
	}
	seen := make(map[T]struct{}, len(src))
	result := make([]T, 0, len(src))
	for _, v := range src {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Difference returns elements in a that are not in b.
//
//	diff := sliceutil.Difference([]int{1,2,3,4}, []int{2,4}) // [1, 3]
func Difference[T comparable](a, b []T) []T {
	bSet := make(map[T]struct{}, len(b))
	for _, v := range b {
		bSet[v] = struct{}{}
	}
	var result []T
	for _, v := range a {
		if _, ok := bSet[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}

// Intersect returns elements that are present in both a and b.
func Intersect[T comparable](a, b []T) []T {
	bSet := make(map[T]struct{}, len(b))
	for _, v := range b {
		bSet[v] = struct{}{}
	}
	var result []T
	for _, v := range a {
		if _, ok := bSet[v]; ok {
			result = append(result, v)
		}
	}
	return Unique(result)
}

// ---------------------------------------------------------------------------
// Convenience
// ---------------------------------------------------------------------------

// ToMap converts a slice to a map using fn to extract the key for each element.
//
//	byID := sliceutil.ToMap(users, func(u User) string { return u.ID })
func ToMap[T any, K comparable](src []T, fn func(T) K) map[K]T {
	result := make(map[K]T, len(src))
	for _, v := range src {
		result[fn(v)] = v
	}
	return result
}

// Keys extracts all keys from a map as a slice.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values extracts all values from a map as a slice.
func Values[K comparable, V any](m map[K]V) []V {
	vals := make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}

// IsEmpty reports whether the slice is nil or has zero length.
func IsEmpty[T any](src []T) bool {
	return len(src) == 0
}
