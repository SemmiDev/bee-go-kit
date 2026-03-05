// Package ptr provides helper functions for converting values to pointers and
// safely dereferencing pointers with defaults.
//
// In Go, creating a pointer to a literal is verbose:
//
//	s := "hello"
//	field.Name = &s
//
// With this package:
//
//	field.Name = ptr.String("hello")
//
// These helpers are especially useful for building structs with optional
// (nullable) fields, which are common in database models and API requests.
package ptr

import "time"

// ---------------------------------------------------------------------------
// Value → Pointer
// ---------------------------------------------------------------------------

// String returns a pointer to the given string value.
func String(v string) *string { return &v }

// Int returns a pointer to the given int value.
func Int(v int) *int { return &v }

// Int32 returns a pointer to the given int32 value.
func Int32(v int32) *int32 { return &v }

// Int64 returns a pointer to the given int64 value.
func Int64(v int64) *int64 { return &v }

// Float32 returns a pointer to the given float32 value.
func Float32(v float32) *float32 { return &v }

// Float64 returns a pointer to the given float64 value.
func Float64(v float64) *float64 { return &v }

// Bool returns a pointer to the given bool value.
func Bool(v bool) *bool { return &v }

// Time returns a pointer to the given time.Time value.
func Time(v time.Time) *time.Time { return &v }

// Of returns a pointer to the given value (generic version).
// Works with any type:
//
//	ptr.Of(42)          // *int
//	ptr.Of("hello")     // *string
//	ptr.Of(MyStruct{})  // *MyStruct
func Of[T any](v T) *T { return &v }

// ---------------------------------------------------------------------------
// Pointer → Value (safe dereference)
// ---------------------------------------------------------------------------

// StringValue returns the value pointed to by p, or empty string if p is nil.
func StringValue(p *string) string {
	if p != nil {
		return *p
	}
	return ""
}

// IntValue returns the value pointed to by p, or 0 if p is nil.
func IntValue(p *int) int {
	if p != nil {
		return *p
	}
	return 0
}

// Int64Value returns the value pointed to by p, or 0 if p is nil.
func Int64Value(p *int64) int64 {
	if p != nil {
		return *p
	}
	return 0
}

// Float64Value returns the value pointed to by p, or 0 if p is nil.
func Float64Value(p *float64) float64 {
	if p != nil {
		return *p
	}
	return 0
}

// BoolValue returns the value pointed to by p, or false if p is nil.
func BoolValue(p *bool) bool {
	if p != nil {
		return *p
	}
	return false
}

// TimeValue returns the value pointed to by p, or zero time if p is nil.
func TimeValue(p *time.Time) time.Time {
	if p != nil {
		return *p
	}
	return time.Time{}
}

// ValueOf returns the value pointed to by p, or the zero value of T if p
// is nil (generic version).
func ValueOf[T any](p *T) T {
	if p != nil {
		return *p
	}
	var zero T
	return zero
}

// ValueOrDefault returns the value pointed to by p, or defaultVal if p is nil.
func ValueOrDefault[T any](p *T, defaultVal T) T {
	if p != nil {
		return *p
	}
	return defaultVal
}

// ---------------------------------------------------------------------------
// Nil checks
// ---------------------------------------------------------------------------

// IsNilOrEmpty returns true if the string pointer is nil or points to an
// empty string.
func IsNilOrEmpty(p *string) bool {
	return p == nil || *p == ""
}
