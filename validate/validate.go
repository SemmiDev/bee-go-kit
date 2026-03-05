// Package validate provides a pre-configured struct validator with Indonesian
// (Bahasa Indonesia) error translations and commonly used custom validation
// rules.
//
// Built on top of github.com/go-playground/validator/v10, it initialises a
// singleton validator instance with:
//   - Indonesian locale translations
//   - Custom validations: uuid_any, base64_pdf, safe_text, base64, base64_image
//   - JSON field-name-based error reporting
//
// Usage:
//
//	// Initialise once (typically in main):
//	validate.Init()
//
//	// Validate a struct:
//	errs := validate.Struct(req)
//	if errs != nil {
//	    // errs is map[string]string — field name → Indonesian error message
//	    httputil.WriteValidationError(w, "Data tidak valid", errs)
//	}
//
//	// Register your own custom validation:
//	validate.RegisterCustom("my_tag", myFunc, "Pesan error kustom")
package validate

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	id_translations "github.com/go-playground/validator/v10/translations/id"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Singleton state
// ---------------------------------------------------------------------------

var (
	instance   *validator.Validate
	translator ut.Translator
	once       sync.Once
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// Config controls how the validator is initialised.
type Config struct {
	// UseIndonesianTranslations enables Bahasa Indonesia error messages.
	// Defaults to true via Init().
	UseIndonesianTranslations bool
}

// ---------------------------------------------------------------------------
// Initialisation
// ---------------------------------------------------------------------------

// Init initialises the singleton validator with Indonesian translations.
// It is safe to call multiple times; only the first call takes effect.
func Init() {
	InitWithConfig(Config{UseIndonesianTranslations: true})
}

// InitWithConfig initialises the singleton validator with custom configuration.
func InitWithConfig(cfg Config) {
	once.Do(func() {
		instance = validator.New()

		// Use JSON tag names for error field paths.
		instance.RegisterTagNameFunc(jsonFieldName)

		if cfg.UseIndonesianTranslations {
			setupIndonesianTranslations()
		}

		registerBuiltinValidations()

		if cfg.UseIndonesianTranslations {
			registerBuiltinTranslations()
			overrideDefaultTranslations()
		}
	})
}

// Reset clears the singleton so Init can be called again (useful for testing).
func Reset() {
	once = sync.Once{}
	instance = nil
	translator = nil
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Validator returns the underlying *validator.Validate instance.
func Validator() *validator.Validate {
	Init()
	return instance
}

// Translator returns the Indonesian translator instance.
func Translator() ut.Translator {
	Init()
	return translator
}

// Struct validates a struct and returns per-field Indonesian error messages.
// Returns nil if validation passes.
func Struct(obj interface{}) map[string]string {
	v := Validator()
	t := Translator()

	if err := v.Struct(obj); err != nil {
		return FormatErrors(err, t)
	}
	return nil
}

// StructWithMessage validates a struct and replaces error messages for specific
// fields. This lets callers override the default Indonesian translations on a
// per-call basis without globally changing them.
//
//	errs := validate.StructWithMessage(req, map[string]string{
//	    "Email": "Alamat email wajib diisi untuk verifikasi",
//	    "Name":  "Nama harus diisi sesuai KTP",
//	})
func StructWithMessage(obj interface{}, customMessages map[string]string) map[string]string {
	errs := Struct(obj)
	if errs == nil || len(customMessages) == 0 {
		return errs
	}
	for field, msg := range customMessages {
		if _, exists := errs[field]; exists {
			errs[field] = msg
		}
	}
	return errs
}

// RegisterCustom registers a custom validation tag with a corresponding
// Indonesian error message. Call this after Init() for project-specific rules.
//
//	validate.RegisterCustom("phone_id", validatePhone, "Nomor telepon tidak valid")
func RegisterCustom(tag string, fn validator.Func, message string) {
	v := Validator()
	_ = v.RegisterValidation(tag, fn)
	registerTranslation(tag, message)
}

// ---------------------------------------------------------------------------
// Built-in custom validations
// ---------------------------------------------------------------------------

func registerBuiltinValidations() {
	_ = instance.RegisterValidation("uuid_any", validateUUIDAnyCase)
	_ = instance.RegisterValidation("base64_pdf", validateBase64PDF)
	_ = instance.RegisterValidation("safe_text", validateSafeText)
	_ = instance.RegisterValidation("base64", validateBase64)
	_ = instance.RegisterValidation("base64_image", validateBase64Image)
}

// validateUUIDAnyCase accepts any valid UUID regardless of case.
func validateUUIDAnyCase(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	if v == "" {
		return true // let 'required' handle empty
	}
	_, err := uuid.Parse(v)
	return err == nil
}

// validateBase64PDF validates "data:application/pdf;base64,<data>".
func validateBase64PDF(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	const prefix = "data:application/pdf;base64,"
	if !strings.HasPrefix(s, prefix) {
		return false
	}
	data := s[len(prefix):]
	if len(data) == 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(data)
	return err == nil
}

// safeTextRegex matches letters, digits, spaces, period, comma, hyphen,
// underscore, and parentheses.
var safeTextRegex = regexp.MustCompile(`^[a-zA-Z0-9 .,_\-\(\)]+$`)

// validateSafeText allows only safe characters in text fields.
func validateSafeText(fl validator.FieldLevel) bool {
	f := fl.Field()
	if f.Kind() != reflect.String {
		return false
	}
	v := f.String()
	if v == "" {
		return true
	}
	return safeTextRegex.MatchString(v)
}

// validateBase64 validates "data:<mime>;base64,<data>".
func validateBase64(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	if !strings.HasPrefix(s, "data:") || !strings.Contains(s, ";base64,") {
		return false
	}
	parts := strings.SplitN(s, ";base64,", 2)
	if len(parts) != 2 || len(parts[1]) == 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(parts[1])
	return err == nil
}

// validateBase64Image validates "data:image/<type>;base64,<data>".
func validateBase64Image(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	if !strings.HasPrefix(s, "data:image/") || !strings.Contains(s, ";base64,") {
		return false
	}
	parts := strings.SplitN(s, ";base64,", 2)
	if len(parts) != 2 || len(parts[1]) == 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(parts[1])
	return err == nil
}

// ---------------------------------------------------------------------------
// Translations
// ---------------------------------------------------------------------------

func setupIndonesianTranslations() {
	loc := id.New()
	uni := ut.New(loc, loc)
	translator, _ = uni.GetTranslator("id")
	_ = id_translations.RegisterDefaultTranslations(instance, translator)
}

func registerBuiltinTranslations() {
	translations := map[string]string{
		"uuid_any":     "Format UUID tidak valid",
		"base64_pdf":   "Format file harus PDF (Base64)",
		"safe_text":    "Hanya boleh berisi huruf, angka, spasi, titik, koma, strip, dan underscore",
		"base64":       "Format Base64 tidak valid",
		"base64_image": "Format gambar tidak valid",
	}
	for tag, msg := range translations {
		registerTranslation(tag, msg)
	}
}

func overrideDefaultTranslations() {
	overrides := map[string]string{
		"required": "Wajib diisi",
		"email":    "Format email tidak valid",
		"numeric":  "Harus berupa angka",
		"min":      "Minimal {0} karakter/digit",
		"max":      "Maksimal {0} karakter/digit",
		"len":      "Harus tepat {0} karakter/digit",
		"gte":      "Harus lebih besar atau sama dengan {0}",
		"lte":      "Harus lebih kecil atau sama dengan {0}",
		"gt":       "Harus lebih besar dari {0}",
		"lt":       "Harus lebih kecil dari {0}",
		"alpha":    "Hanya boleh berisi huruf",
		"alphanum": "Hanya boleh berisi huruf dan angka",
		"url":      "Format URL tidak valid",
		"oneof":    "Harus salah satu dari {0}",
	}
	for tag, msg := range overrides {
		registerTranslation(tag, msg)
	}
}

func registerTranslation(tag, message string) {
	_ = instance.RegisterTranslation(tag, translator,
		func(ut ut.Translator) error {
			return ut.Add(tag, message, true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			msg := message
			if fe.Param() != "" {
				msg = strings.Replace(msg, "{0}", fe.Param(), -1)
			}
			return msg
		},
	)
}

// ---------------------------------------------------------------------------
// Error formatting
// ---------------------------------------------------------------------------

// FormatErrors converts validator.ValidationErrors into a map of field → message.
// It is exported so clients can use their own ut.Translator.
func FormatErrors(err error, trans ut.Translator) map[string]string {
	result := make(map[string]string)

	valErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return result
	}

	for _, fe := range valErrs {
		fieldPath := cleanFieldPath(fe.Namespace())
		if _, exists := result[fieldPath]; exists {
			continue // keep the first error per field
		}
		result[fieldPath] = errorMessage(fe, trans)
	}

	return result
}

func errorMessage(fe validator.FieldError, trans ut.Translator) string {
	if trans != nil {
		if msg := fe.Translate(trans); msg != "" {
			return msg
		}
	}
	return fallbackMessage(fe)
}

// fallbackMessage provides hardcoded Indonesian messages when the translator
// is not available.
func fallbackMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Wajib diisi"
	case "email":
		return "Format email tidak valid"
	case "uuid_any":
		return "Format UUID tidak valid"
	case "numeric":
		return "Harus berupa angka"
	case "base64_pdf":
		return "Format file harus PDF (Base64)"
	case "min":
		return fmt.Sprintf("Minimal %s karakter/digit", fe.Param())
	case "max":
		return fmt.Sprintf("Maksimal %s karakter/digit", fe.Param())
	case "len":
		return fmt.Sprintf("Harus tepat %s karakter/digit", fe.Param())
	case "gte":
		return fmt.Sprintf("Harus lebih besar atau sama dengan %s", fe.Param())
	case "lte":
		return fmt.Sprintf("Harus lebih kecil atau sama dengan %s", fe.Param())
	case "gt":
		return fmt.Sprintf("Harus lebih besar dari %s", fe.Param())
	case "lt":
		return fmt.Sprintf("Harus lebih kecil dari %s", fe.Param())
	case "alpha":
		return "Hanya boleh berisi huruf"
	case "alphanum":
		return "Hanya boleh berisi huruf dan angka"
	case "url":
		return "Format URL tidak valid"
	case "safe_text":
		return "Hanya boleh berisi huruf, angka, spasi, titik, koma, strip, dan underscore"
	case "base64":
		return "Format Base64 tidak valid"
	case "base64_image":
		return "Format gambar tidak valid"
	default:
		return fmt.Sprintf("Validasi gagal untuk aturan '%s'", fe.Tag())
	}
}

// cleanFieldPath removes the root struct name from the namespace.
//
//	"User.Email"        → "Email"
//	"User.Address.City" → "Address.City"
func cleanFieldPath(namespace string) string {
	parts := strings.SplitN(namespace, ".", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return namespace
}

// jsonFieldName extracts the JSON tag name from a struct field.
func jsonFieldName(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}
