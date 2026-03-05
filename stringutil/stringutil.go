// Package stringutil provides common string manipulation utilities for API
// development: truncation, masking, slugification, and more.
package stringutil

import (
	"regexp"
	"strings"
	"unicode"
)

// ---------------------------------------------------------------------------
// Truncation
// ---------------------------------------------------------------------------

// Truncate shortens s to maxLen characters and appends "..." if it was
// truncated. Returns s unchanged if it fits within maxLen.
//
//	stringutil.Truncate("Hello World", 5) → "Hello..."
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// TruncateWithSuffix truncates s and appends the custom suffix.
func TruncateWithSuffix(s string, maxLen int, suffix string) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + suffix
}

// ---------------------------------------------------------------------------
// Masking
// ---------------------------------------------------------------------------

// MaskEmail masks the local part of an email address for privacy. Shows the
// first and last character of the local part, replacing the rest with '*'.
//
//	stringutil.MaskEmail("john.doe@example.com") → "j******e@example.com"
//	stringutil.MaskEmail("ab@x.com")             → "a*b@x.com"
func MaskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || len(parts[0]) == 0 {
		return email
	}
	local := parts[0]
	domain := parts[1]

	runes := []rune(local)
	if len(runes) <= 2 {
		return string(runes[0]) + "*" + "@" + domain
	}
	masked := string(runes[0]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1])
	return masked + "@" + domain
}

// MaskPhone masks a phone number, showing only the last 4 digits.
//
//	stringutil.MaskPhone("+6281234567890") → "**********7890"
func MaskPhone(phone string) string {
	runes := []rune(phone)
	if len(runes) <= 4 {
		return phone
	}
	return strings.Repeat("*", len(runes)-4) + string(runes[len(runes)-4:])
}

// ---------------------------------------------------------------------------
// Slug
// ---------------------------------------------------------------------------

// nonAlphanumRegex matches non-alphanumeric characters.
var nonAlphanumRegex = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a string to a URL-friendly slug.
//
//	stringutil.Slugify("Hello World!") → "hello-world"
//	stringutil.Slugify("  Foo BAR  ") → "foo-bar"
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphanumRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// ---------------------------------------------------------------------------
// Case conversion
// ---------------------------------------------------------------------------

// ToSnakeCase converts CamelCase or PascalCase to snake_case.
//
//	stringutil.ToSnakeCase("MyFieldName") → "my_field_name"
//	stringutil.ToSnakeCase("HTTPStatusCode") → "http_status_code"
func ToSnakeCase(s string) string {
	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			// Insert underscore before uppercase letter unless:
			// - it's the first character
			// - previous char is already uppercase AND next char is also uppercase (acronym middle)
			if i > 0 {
				prev := runes[i-1]
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					result.WriteRune('_')
				} else if unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ---------------------------------------------------------------------------
// Misc
// ---------------------------------------------------------------------------

// Coalesce returns the first non-empty string from the given values.
// Returns empty string if all values are empty.
//
//	name := stringutil.Coalesce(nickname, fullName, "Anonymous")
func Coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// ContainsAny reports whether s contains any of the given substrings
// (case-insensitive).
func ContainsAny(s string, substrs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

// DefaultIfEmpty returns s if it is non-empty, otherwise returns defaultVal.
func DefaultIfEmpty(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}

// Reverse returns the string with its characters in reverse order.
// Handles Unicode correctly.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
