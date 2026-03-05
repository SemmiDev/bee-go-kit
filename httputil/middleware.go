package httputil

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/google/uuid"
	"github.com/semmidev/bee-go-kit/apperror"
)

// ---------------------------------------------------------------------------
// Recovery middleware — catches panics and returns 500 JSON
// ---------------------------------------------------------------------------

// Recovery returns an http.Handler middleware that recovers from panics,
// logs the stack trace, and returns a 500 JSON error response instead of
// crashing the server.
//
// Usage with standard ServeMux:
//
//	mux := http.NewServeMux()
//	handler := httputil.Recovery(mux)
//	http.ListenAndServe(":8080", handler)
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Log the panic with full stack trace.
				slog.Error("panic recovered",
					"panic", fmt.Sprint(rec),
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)

				// Respond with a structured 500 error.
				WriteError(w, apperror.Internal(
					"internal server error",
					fmt.Errorf("panic: %v", rec),
				))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ---------------------------------------------------------------------------
// Request ID middleware — injects a unique request ID
// ---------------------------------------------------------------------------

// requestIDHeader is the HTTP header key for request IDs.
const requestIDHeader = "X-Request-Id"

// RequestID returns an http.Handler middleware that ensures every request has
// a unique X-Request-Id header. If the incoming request already has one, it
// is preserved; otherwise a new UUID v4 is generated.
//
// The request ID is also set on the response headers for correlation.
//
// Usage:
//
//	handler := httputil.RequestID(httputil.Recovery(mux))
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = uuid.New().String()
			r.Header.Set(requestIDHeader, id)
		}
		// Echo the request ID back in the response.
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r)
	})
}

// GetRequestID extracts the request ID from the request headers.
// Returns empty string if no request ID was set.
func GetRequestID(r *http.Request) string {
	return r.Header.Get(requestIDHeader)
}

// ---------------------------------------------------------------------------
// No-cache middleware — prevents browser/proxy caching of API responses
// ---------------------------------------------------------------------------

// NoCache returns an http.Handler middleware that sets headers to prevent
// caching of API responses. Useful for JSON API endpoints.
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}
