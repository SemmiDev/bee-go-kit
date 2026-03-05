# bee-go-kit

A curated collection of reusable Go packages extracted from production API services. Each package is self-contained, framework-agnostic (where possible), and fully documented with godoc comments.

## Packages

| Package | Description |
|---|---|
| **`apperror`** | Structured HTTP errors with status codes, error codes, and checker utilities |
| **`httputil`** | API response envelope, JSON writers, HTTP error handlers, and middleware (recovery, request ID, no-cache) |
| **`pagination`** | Filter/pagination parsing from HTTP requests and paginated response metadata |
| **`sorting`** | SQL-injection-safe ORDER BY builder with configurable column quoting (MSSQL/PG/MySQL) |
| **`sanitize`** | Input sanitisation for strings, emails, filenames, and SQL LIKE patterns |
| **`dbx`** | Context-aware transaction manager and unified DB executor for sqlx |
| **`cache`** | Cache interface + production-ready Redis implementation (`cache/redis`) |
| **`storage`** | Storage interface + S3 proxy implementation (`storage/s3`) |
| **`validate`** | Struct validator with Indonesian (Bahasa Indonesia) translations and extensible rules |
| **`netutil`** | Network utility (local IP detection) |
| **`ptr`** | Pointer helpers — value-to-pointer + safe dereference (with generics) |
| **`sliceutil`** | Generic slice utilities — Map, Filter, Reduce, Contains, Unique, Chunk, GroupBy, ToMap |
| **`stringutil`** | String utilities — truncation, email/phone masking, slugify, snake_case, coalesce |
| **`timeutil`** | Time utilities — Indonesian locale (WIB/WITA/WIT), date boundaries, parsing, comparison |

## Quick Start

```go
import (
    "github.com/semmidev/bee-go-kit/apperror"
    "github.com/semmidev/bee-go-kit/httputil"
    "github.com/semmidev/bee-go-kit/pagination"
    "github.com/semmidev/bee-go-kit/validate"
    "github.com/semmidev/bee-go-kit/ptr"
    "github.com/semmidev/bee-go-kit/sliceutil"
    "github.com/semmidev/bee-go-kit/timeutil"
)

// Validate a request
validate.Init()
errs := validate.Struct(req)
if errs != nil {
    httputil.WriteValidationError(w, "Data tidak valid", errs)
    return
}

// Return paginated data
paging, _ := pagination.NewPaging(filter.CurrentPage, filter.PerPage, totalData)
httputil.WriteSuccessWithPaging(w, http.StatusOK, "Success", data, paging)

// Return error
httputil.WriteError(w, apperror.NotFound("User not found"))

// Pointer helpers (for optional DB fields)
user.Name = ptr.String("John")
name := ptr.ValueOrDefault(user.Name, "Unknown")

// Slice utilities
ids := sliceutil.Map(users, func(u User) string { return u.ID })
admins := sliceutil.Filter(users, func(u User) bool { return u.Role == "admin" })
batches := sliceutil.Chunk(allItems, 100)

// Indonesian time
fmt.Println(timeutil.FormatIndonesian(time.Now())) // "5 Maret 2026, 14:01 WIB"

// Middleware stack
handler := httputil.RequestID(httputil.Recovery(mux))
```

## License
