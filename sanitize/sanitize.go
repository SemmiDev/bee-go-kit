// Package sanitize provides input sanitisation utilities for strings,
// emails, filenames, and SQL LIKE patterns.
//
// These functions are designed to be used at the boundary between user input
// and application logic to prevent XSS, SQL injection via LIKE wildcards,
// path traversal, and other common input-based attacks.
//
// For SQL parameterised queries, prefer SanitizeStringNoEscape (no HTML
// escaping needed) over SanitizeString (which HTML-escapes).
package sanitize

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

// ---------------------------------------------------------------------------
// Composable sanitiser pipeline
// ---------------------------------------------------------------------------

// StringFunc is a function that transforms a string. Use it to build custom
// sanitisation pipelines with Chain or Apply.
//
//	myPipeline := sanitize.Chain(
//	    sanitize.String,
//	    strings.ToLower,
//	    myCustomTrimmer,
//	)
//	clean := myPipeline("  RAW INPUT  ")
type StringFunc func(string) string

// Chain composes multiple StringFunc into a single one. Functions are applied
// in order from left to right.
//
//	pipeline := sanitize.Chain(sanitize.String, strings.ToLower)
//	pipeline("<b>Hello</b>  ") // → "&lt;b&gt;hello&lt;/b&gt;"
func Chain(fns ...StringFunc) StringFunc {
	return func(s string) string {
		for _, fn := range fns {
			s = fn(s)
		}
		return s
	}
}

// Apply runs a string through the given sanitiser functions in order.
//
//	clean := sanitize.Apply(raw, sanitize.String, strings.ToLower)
func Apply(s string, fns ...StringFunc) string {
	for _, fn := range fns {
		s = fn(s)
	}
	return s
}

// ---------------------------------------------------------------------------
// Built-in sanitisers
// ---------------------------------------------------------------------------

// String sanitises a string value by trimming surrounding whitespace and
// escaping HTML special characters to prevent XSS.
func String(s string) string {
	return html.EscapeString(strings.TrimSpace(s))
}

// StringNoEscape trims whitespace but preserves original characters.
// Use when HTML escaping is not needed (e.g. parameterised DB queries).
func StringNoEscape(s string) string {
	return strings.TrimSpace(s)
}

// ForSQLLike escapes SQL LIKE wildcard characters (%, _, [) so user input
// is treated as literal text, not as pattern metacharacters.
//
//	sanitize.ForSQLLike("100%")  → "100[%]"
//	sanitize.ForSQLLike("a_b")  → "a[_]b"
func ForSQLLike(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "[", "[[]")
	s = strings.ReplaceAll(s, "%", "[%]")
	s = strings.ReplaceAll(s, "_", "[_]")
	return s
}

// emailRegex is compiled once at init time for performance.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email normalises an email address (trim + lowercase) and validates its
// format. Returns empty string if the email is invalid.
func Email(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	if !emailRegex.MatchString(email) {
		return ""
	}
	return email
}

// Alphanumeric removes all characters that are not letters or digits.
//
//	sanitize.Alphanumeric("abc-123!") → "abc123"
func Alphanumeric(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Numeric removes all non-digit characters.
//
//	sanitize.Numeric("+62-812-3456") → "628123456"
func Numeric(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Filename removes null bytes, path traversal sequences, and characters that
// are dangerous in file names across operating systems.
func Filename(filename string) string {
	// Remove null bytes and path traversal.
	filename = strings.ReplaceAll(filename, "\x00", "")
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")

	// Remove characters that are problematic on Windows / URLs.
	for _, ch := range []string{"<", ">", ":", "\"", "|", "?", "*"} {
		filename = strings.ReplaceAll(filename, ch, "")
	}

	return strings.TrimSpace(filename)
}

// Ptr sanitises a pointer string using StringNoEscape. Returns nil if the
// result is empty.
func Ptr(s *string) *string {
	if s == nil {
		return nil
	}
	result := StringNoEscape(*s)
	if result == "" {
		return nil
	}
	return &result
}
