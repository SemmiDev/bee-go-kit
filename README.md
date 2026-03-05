# 🐝 bee-go-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/semmidev/bee-go-kit.svg)](https://pkg.go.dev/github.com/semmidev/bee-go-kit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A curated, highly-extensible collection of reusable Go packages extracted from production API services. Each package is self-contained, framework-agnostic, and designed following SOLID principles.

`bee-go-kit` provides a robust foundation for building modern web APIs in Go, offering structured error handling, standardized responses, pagination, safe sorting, composable sanitization, and pre-configured validation with Indonesian (Bahasa Indonesia) translations.

---

## 📦 Installation

```bash
go get github.com/semmidev/bee-go-kit
```

---

## 🚀 Core Packages: Usage & Extensibility

Every package in `bee-go-kit` is designed to be highly extensible. You can completely customize the behavior without forking the repository.

### 1. `validate` (Flexible Struct Validation)
A pre-configured struct validator using `go-playground/validator/v10` with Indonesian error messages.

**Basic Usage:**
```go
import "github.com/semmidev/bee-go-kit/validate"

validate.Init() // Call once at app startup

type User struct {
    Email string `validate:"required,email"`
    Age   int    `validate:"gte=18"`
}

errs := validate.Struct(User{Email: "invalid"})
// errs: {"Email": "Format email tidak valid", "Age": "Harus lebih besar atau sama dengan 18"}
```

**Extending / Customizing:**
```go
// 1. Add entirely new validation rules with custom Indonesian messages
validate.RegisterCustom("is_adult", func(fl validator.FieldLevel) bool {
    return fl.Field().Int() >= 18
}, "Pengguna harus berusia minimal 18 tahun")

// 2. Override an error message for a SPECIFIC HTTP request
//    (Useful when standard translations aren't context-aware enough)
errs := validate.StructWithMessage(req, map[string]string{
    "Email": "Alamat email wajib diisi untuk registrasi akun Anda",
})

// 3. Use your own custom translator (e.g. for English or another language)
myTranslator := setupEnglishTranslator()
errs := validate.FormatErrors(err, myTranslator)
```

### 2. `apperror` (Structured HTTP Errors)
Structured HTTP errors mapped to stable, machine-readable error codes.

**Basic Usage:**
```go
import "github.com/semmidev/bee-go-kit/apperror"

// Built-in constructors
err := apperror.NotFound("User not found")
err := apperror.BadRequest("Invalid input data")

// Checking errors
if apperror.IsNotFound(err) { /* Handle 404 */ }
status := apperror.HTTPStatus(err) // Safely extract HTTP status code from any error
```

**Extending / Customizing:**
```go
// 1. Create fully custom HTTP errors with your own Status Codes and Error Codes
err := apperror.New(http.StatusPaymentRequired, "ERR_NO_CREDIT", "Please top up your balance")

// 2. Formatted custom errors
err = apperror.Newf(http.StatusConflict, "ERR_DUPLICATE", "User %s already exists", email)

// 3. Wrap underlying database or internal errors for debugging,
//    but keep the HTTP response clean
err = apperror.NewWithErr(http.StatusBadGateway, "ERR_UPSTREAM", "Failed to reach payment gateway", actualErr)
```

### 3. `httputil` (Responses & Middleware)
Standardized JSON responses and middleware.

**Basic Usage:**
```go
import "github.com/semmidev/bee-go-kit/httputil"

// 1. Success response (200 OK)
httputil.WriteSuccess(w, http.StatusOK, "User created", userObj)

// 2. Error response (automatically extracts status code from apperror)
httputil.WriteError(w, apperror.NotFound("Data missing"))

// 3. Middlewares (Panic recovery, Request IDs, No-Cache headers)
handler := httputil.RequestID(httputil.Recovery(mux))
```

**Extending / Customizing:**
```go
// Handling Domain or Framework-specific errors:
// If your application uses custom error structures (e.g., gRPC errors, Domain errors),
// you can teach httputil how to format them using the ErrorTransformer hook.
httputil.SetErrorTransformer(func(err error) (int, httputil.Response) {
    if myErr, ok := err.(*domain.Error); ok {
        return myErr.Status, httputil.Response{
            Success:   false,
            ErrorCode: myErr.Code,
            Message:   myErr.Message,
        }
    }
    // Return 0 to fall back to the default bee-go-kit error handling
    return 0, httputil.Response{}
})
```

### 4. `pagination` (Safe Filtering & Paging)
Extracts array/list parameters safely from HTTP requests.

**Basic Usage:**
```go
import "github.com/semmidev/bee-go-kit/pagination"

// Parses params like ?page=2&per_page=20&keyword=john
filter := pagination.FilterFromRequest(r)

// Safe offset variables for SQL
sqlOffset := filter.GetOffset()
sqlLimit := filter.GetLimit()

// Send paginated response
pagingInfo, _ := pagination.NewPaging(filter.CurrentPage, filter.PerPage, totalRecords)
httputil.WriteSuccessWithPaging(w, 200, "Success", records, pagingInfo)
```

**Extending / Customizing:**
```go
// Adapt to custom Frontend query param names without breaking the library!
// Example: If your frontend sends ?pageNo=1&pageSize=10 instead of ?page=1
params := pagination.DefaultParamNames()
params.Page = "pageNo"
params.PerPage = "pageSize"
params.Keyword = "searchQuery"

filter := pagination.FilterFromRequestWithParams(r, params)
// Now `filter` contains the correct values extracted from the custom keys
```

### 5. `sorting` (Safe SQL Order By)
Prevents SQL injection when binding user inputs to `ORDER BY` clauses.

**Basic Usage:**
```go
import "github.com/semmidev/bee-go-kit/sorting"

// Whitelist mapping (Frontend Sort Key -> Actual DB Column)
cfg := sorting.NewSortConfig(map[string]string{
    "createdAt": "created_date",
    "userName":  "user_name",
}, "created_date")

// Automatically protects against injection
clause := sorting.BuildFullOrderByClause("userName", "desc", cfg)
// Output: ORDER BY user_name DESC
```

**Extending / Customizing:**
```go
// Customize the SQL quote character depending on your database engine
cfg = cfg.WithQuote(sorting.PostgresQuote) // Output: ORDER BY "user_name" DESC
cfg = cfg.WithQuote(sorting.MySQLQuote)    // Output: ORDER BY `user_name` DESC
cfg = cfg.WithQuote(sorting.MSSQLQuote)    // Output: ORDER BY [user_name] DESC

// Supply a completely custom quote function for proprietary SQL engines
cfg = cfg.WithQuote(func(col string) string { return fmt.Sprintf("'%s'", col) })
```

### 6. `sanitize` (Input Sanitization Pipelines)
Prevent XSS, Path Traversal, and SQL injection.

**Basic Usage:**
```go
import "github.com/semmidev/bee-go-kit/sanitize"

safeHTML := sanitize.String("<script>alert(1)</script>") // Escapes HTML
safeLike := sanitize.ForSQLLike("100%")                 // Escapes % to [%] for SQL LIKE
cleanEmail := sanitize.Email("  User@Email.com ")       // -> user@email.com
safeFile := sanitize.Filename("../../../etc/passwd")    // -> etcpasswd
```

**Extending / Customizing:**
```go
// Build powerful, composable sanitization pipelines that merge
// bee-go-kit's sanitizers with your own custom logic!
myPipeline := sanitize.Chain(
    sanitize.StringNoEscape,                                     // 1. Trim whitespace
    strings.ToLower,                                             // 2. Make lowercase
    func(s string) string { return strings.ReplaceAll(s, "bad", "***") }, // 3. Custom filter
)

cleanString := myPipeline("   BAD INPUT   ") // Result: "*** input"
```

---

## 🧰 Adapters & Data Utilities

### Interfaces (`dbx`, `cache`, `storage`)
The library uses robust Go interfaces to remain entirely framework agnostic:
* **`dbx`**: Implements an `Executor` interface that transparently wraps both standard queries (`*sqlx.DB`) and transactions (`*sqlx.Tx`). It includes a complete `TransactionManager`.
* **`cache`**: A generic `Cache` interface allowing you to plug in Redis, Memcached, etc. A production-ready **Redis implementation** (`cache/redis`) is included.
* **`storage`**: A generic `Storage` interface for uploading/downloading files. An **S3 implementation** (`storage/s3`) is included.

### Tooling Packages

| Package | Purpose & Examples |
|---------|---------------------|
| **`sliceutil`** | Generics for slices.<br/>- `sliceutil.Map(users, func(u User) string { return u.ID })`<br/>- `sliceutil.Filter(...)`, `Reduce`, `GroupBy`, `Chunk`, `Unique` |
| **`ptr`** | Generics for pointers (handling optional struct fields).<br/>- `user.Age = ptr.Of(25)`<br/>- `age := ptr.ValueOrDefault(user.Age, 18)` |
| **`stringutil`**| String adjustments.<br/>- `Truncate("very long string", 10)`<br/>- `MaskEmail("john@example.com")` -> `j***@example.com`<br/>- `ToSnakeCase`, `Slugify`, `Coalesce` |
| **`timeutil`** | Time manipulation.<br/>- `FormatIndonesian(time.Now())` -> `"5 Maret 2026, 14:01 WIB"`<br/>- `StartOfDay(t)`, `EndOfDay(t)`, `IsBetween(t, start, end)` |
| **`netutil`** | Networking helpers.<br/>- `MustGetIP()` -> Retrieves local non-loopback IPv4 |

---

## ⚙️ Commands & Development

This repository includes a `Makefile` to quickly run tests, format code, and check coverage. All packages include extensive BDD (Behavior-Driven Development) tests using `GoConvey`.

| Command | Description |
|---------|-------------|
| `make test` | Run all BDD tests |
| `make test-v` | Run tests with verbose GoConvey output |
| `make test-cover` | Run tests and print coverage stats |
| `make cover-html` | Generate and open HTML coverage report |
| `make lint` | Run go vet & staticcheck |
| `make doc` | Serve local `pkgsite` documentation on `http://localhost:8080` |

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
