# 🐝 bee-go-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/semmidev/bee-go-kit.svg)](https://pkg.go.dev/github.com/semmidev/bee-go-kit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A curated, highly-extensible collection of reusable Go packages extracted from production API services. Each package is self-contained, framework-agnostic, and designed following SOLID principles.

`bee-go-kit` provides a robust foundation for building modern web APIs in Go, offering structured error handling, standardized responses, pagination, safe sorting, composable sanitization, and pre-configured validation with Indonesian (Bahasa Indonesia) translations.

---

## 📑 Table of Contents
- [🐝 bee-go-kit](#-bee-go-kit)
  - [📑 Table of Contents](#-table-of-contents)
  - [📦 Installation](#-installation)
  - [🧠 Design Philosophy](#-design-philosophy)
  - [🚀 Core Packages \& Usage](#-core-packages--usage)
    - [1. `validate` (Validation)](#1-validate-validation)
    - [2. `apperror` (Error Handling)](#2-apperror-error-handling)
    - [3. `httputil` (Responses \& Middleware)](#3-httputil-responses--middleware)
    - [4. `pagination` (Pagination \& Filtering)](#4-pagination-pagination--filtering)
    - [5. `sorting` (Safe SQL Order By)](#5-sorting-safe-sql-order-by)
    - [6. `sanitize` (Input Sanitization)](#6-sanitize-input-sanitization)
  - [🧰 Data Utilities](#-data-utilities)
  - [🔌 Interfaces \& Abstractions](#-interfaces--abstractions)
  - [🛠 Extending the Library](#-extending-the-library)
    - [1. Extending Validation (`validate`)](#1-extending-validation-validate)
    - [2. Custom HTTP Errors (`apperror` \& `httputil`)](#2-custom-http-errors-apperror--httputil)
    - [3. Custom Pagination Keys (`pagination`)](#3-custom-pagination-keys-pagination)
    - [4. Custom Sanitization Pipelines (`sanitize`)](#4-custom-sanitization-pipelines-sanitize)
  - [⚙️ Commands \& Development](#️-commands--development)
  - [📄 License](#-license)

---

## 📦 Installation

```bash
go get github.com/semmidev/bee-go-kit
```

---

## 🧠 Design Philosophy
1. **Framework Agnostic**: Works seamlessly with `net/http`, Echo, Gin, Fiber, or Beego.
2. **Highly Extensible**: Almost every behavior can be extended or overridden without forking the codebase.
3. **Fail-safe Defaults**: Provides sane defaults out of the box (e.g., SQL injection prevention in sorting).
4. **Consistency**: Forces a consistent API response contract across your entire ecosystem.

---

## 🚀 Core Packages & Usage

### 1. `validate` (Validation)
A pre-configured struct validator using `go-playground/validator/v10` with **Indonesian error messages**.

```go
import "github.com/semmidev/bee-go-kit/validate"

// Initialize once (e.g., in main.go)
validate.Init()

type User struct {
    Email string `validate:"required,email"`
    Age   int    `validate:"gte=18"`
}

errs := validate.Struct(User{Email: "invalid"})
// errs maps field names to Indonesian messages:
// {"Email": "Format email tidak valid", "Age": "Harus lebih besar atau sama dengan 18"}
```

### 2. `apperror` (Error Handling)
Structured HTTP errors with status codes and stable machine-readable error codes.

```go
import "github.com/semmidev/bee-go-kit/apperror"

// Built-in constructors
err := apperror.NotFound("User not found")
err := apperror.BadRequest("Invalid input data")
err := apperror.Conflict("Email already exists")

// Error checking
if apperror.IsNotFound(err) {
    // Handle 404
}

// Extract HTTP status from ANY error
status := apperror.HTTPStatus(err)
```

### 3. `httputil` (Responses & Middleware)
Standardized JSON responses and common HTTP middlewares.

**Standard Response Wrapper:**
```go
import "github.com/semmidev/bee-go-kit/httputil"

// Success (200 OK)
httputil.WriteSuccess(w, http.StatusOK, "User created", userObj)
/*
{
  "success": true,
  "message": "User created",
  "data": { "id": 1, "name": "John" }
}
*/

// Errors automatically map `apperror` to the correct HTTP status
httputil.WriteError(w, apperror.NotFound("Data missing"))
/*
{
  "success": false,
  "error_code": "ERR_NOT_FOUND",
  "message": "Data missing"
}
*/
```

**Middlewares:**
```go
mux := http.NewServeMux()
// Recovers from panics, adds X-Request-Id, prevents caching
handler := httputil.NoCache(httputil.RequestID(httputil.Recovery(mux)))
```

### 4. `pagination` (Pagination & Filtering)
Extracts list parameters from HTTP requests safely.

```go
import "github.com/semmidev/bee-go-kit/pagination"

filter := pagination.FilterFromRequest(r) // parses ?page=2&per_page=20&keyword=john

// Use in SQL queries safely
offset := filter.GetOffset()
limit := filter.GetLimit()

// Return response with paging metadata
paging, _ := pagination.NewPaging(filter.CurrentPage, filter.PerPage, totalRecords)
httputil.WriteSuccessWithPaging(w, 200, "Success", dataList, paging)
```

### 5. `sorting` (Safe SQL Order By)
Prevents SQL injection when binding user-provided sort columns to DB queries.

```go
import "github.com/semmidev/bee-go-kit/sorting"

// Whitelist allowed columns (Client Key -> DB Column)
cfg := sorting.NewSortConfig(map[string]string{
    "createdAt": "created_date",
    "name":      "user_name",
}, "created_date").WithQuote(sorting.PostgresQuote)

// Builds safely: ORDER BY "user_name" DESC
clause := sorting.BuildFullOrderByClause("name", "desc", cfg)
```

### 6. `sanitize` (Input Sanitization)
Prevent XSS, Path Traversal, and SQL LIKE injection.

```go
import "github.com/semmidev/bee-go-kit/sanitize"

safeHTML := sanitize.String("<script>alert(1)</script>") // Escapes HTML
safeLike := sanitize.ForSQLLike("100%")                 // Escapes % to [%] for SQL
cleanEmail := sanitize.Email("  User@Email.com ")       // -> user@email.com
safeFile := sanitize.Filename("../../../etc/passwd")    // -> etcpasswd
```

---

## 🧰 Data Utilities

| Package | Purpose & Highlight |
|---------|---------------------|
| `sliceutil` | Generics for slices: `Map`, `Filter`, `Reduce`, `GroupBy`, `Chunk`, `Unique` |
| `ptr` | Generics for pointers: `ptr.Of("value")`, `ptr.ValueOrDefault(p, "fallback")` |
| `stringutil`| `Truncate`, `MaskEmail` (`a***@example.com`), `ToSnakeCase`, `Slugify` |
| `timeutil` | Indonesian Dates (`FormatIndonesian`), standard boundaries (`StartOfDay`) |
| `netutil` | `MustGetIP()` (Retrieves local non-loopback IPv4) |

---

## 🔌 Interfaces & Abstractions

`bee-go-kit` provides standard interfaces to ensure clean architecture across your microservices:

- **`dbx`**: `Executor` interface masking both `*sqlx.DB` and `*sqlx.Tx`. Includes a Context-aware `TransactionManager`.
- **`cache`**: Standard `Cache` interface with a production-ready `cache/redis` implementation provided.
- **`storage`**: Standard `Storage` interface (Upload/Download) with an S3 proxy implementation provided.

---

## 🛠 Extending the Library

`bee-go-kit` is built to be extended. You do not need to fork the library to change its behavior.

### 1. Extending Validation (`validate`)
Register your own complex domain validations or override error messages.
```go
// Add a custom validation rule
validate.RegisterCustom("is_adult", func(fl validator.FieldLevel) bool {
    return fl.Field().Int() >= 18
}, "Pengguna harus berusia minimal 18 tahun")

// Override an error message for a specific HTTP request
errs := validate.StructWithMessage(req, map[string]string{
    "Email": "Alamat email wajib diisi untuk registrasi",
})
```

### 2. Custom HTTP Errors (`apperror` & `httputil`)
Have a domain-specific error format (like gRPC errors)? Use `ErrorTransformer`:

```go
httputil.SetErrorTransformer(func(err error) (int, httputil.Response) {
    if domainErr, ok := err.(*MyDomainError); ok {
        return domainErr.HTTPStatus, httputil.Response{
            Success:   false,
            ErrorCode: domainErr.Code,
            Message:   domainErr.Message,
        }
    }
    return 0, httputil.Response{} // Fallback to default bee-go-kit logic
})
```
You can also build custom app errors not included in the defaults:
```go
err := apperror.New(http.StatusPaymentRequired, "ERR_PAYMENT", "Payment required")
```

### 3. Custom Pagination Keys (`pagination`)
If your frontend sends `?pageNo=1&pageSize=10` instead of `?page=1&per_page=10`, adapt seamlessly:
```go
params := pagination.DefaultParamNames()
params.Page = "pageNo"
params.PerPage = "pageSize"
filter := pagination.FilterFromRequestWithParams(r, params)
```

### 4. Custom Sanitization Pipelines (`sanitize`)
Chain built-in sanitizers with your own proprietary clean-up functions:
```go
myPipeline := sanitize.Chain(
    sanitize.StringNoEscape,
    strings.ToLower,
    func(s string) string { return strings.ReplaceAll(s, "badword", "***") },
)
cleanString := myPipeline("   BADWORD   ") // output: "***"
```

---

## ⚙️ Commands & Development

This repository includes a `Makefile` to quickly run tests, format code, and check coverage.

| Command | Description |
|---------|-------------|
| `make test` | Run all BDD tests via GoConvey |
| `make test-cover` | Run tests and print coverage stats |
| `make cover-html` | Generate and open HTML coverage report |
| `make lint` | Run go vet & staticcheck |
| `make doc` | Serve local `pkgsite` documentation on `http://localhost:8080` |

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
