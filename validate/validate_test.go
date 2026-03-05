package validate_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-playground/validator/v10"
	"github.com/semmidev/bee-go-kit/validate"
)

// ---------------------------------------------------------------------------
// Setup helper
// ---------------------------------------------------------------------------

func resetAndInit() {
	validate.Reset()
	validate.Init()
}

// ---------------------------------------------------------------------------
// Init / Singleton
// ---------------------------------------------------------------------------

func TestInit(t *testing.T) {
	Convey("Given Init", t, func() {
		resetAndInit()

		Convey("Then Validator should be non-nil", func() {
			So(validate.Validator(), ShouldNotBeNil)
		})

		Convey("Then Translator should be non-nil", func() {
			So(validate.Translator(), ShouldNotBeNil)
		})
	})
}

func TestInitWithConfig(t *testing.T) {
	Convey("Given InitWithConfig with Indonesian translations disabled", t, func() {
		validate.Reset()
		validate.InitWithConfig(validate.Config{UseIndonesianTranslations: false})

		Convey("Then Validator should still be non-nil", func() {
			So(validate.Validator(), ShouldNotBeNil)
		})
	})
}

func TestReset(t *testing.T) {
	Convey("Given Reset", t, func() {
		resetAndInit()
		v1 := validate.Validator()

		validate.Reset()
		validate.Init()
		v2 := validate.Validator()

		Convey("Then a new validator instance should be created", func() {
			// After reset + re-init, the pointer address should differ.
			So(v1, ShouldNotEqual, v2)
		})
	})
}

// ---------------------------------------------------------------------------
// Struct validation
// ---------------------------------------------------------------------------

func TestStruct(t *testing.T) {
	Convey("Given the Struct validator", t, func() {
		resetAndInit()

		type User struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
			Age   int    `validate:"required,min=1,max=150"`
		}

		Convey("When struct is valid", func() {
			u := User{Name: "Alice", Email: "alice@example.com", Age: 30}
			errs := validate.Struct(u)

			Convey("Then it should return nil", func() {
				So(errs, ShouldBeNil)
			})
		})

		Convey("When struct has missing required fields", func() {
			u := User{}
			errs := validate.Struct(u)

			Convey("Then it should return errors for each field", func() {
				So(errs, ShouldNotBeNil)
				So(errs, ShouldContainKey, "Name")
				So(errs, ShouldContainKey, "Email")
			})
		})

		Convey("When email is invalid", func() {
			u := User{Name: "Bob", Email: "not-an-email", Age: 25}
			errs := validate.Struct(u)

			Convey("Then it should return an error for Email", func() {
				So(errs, ShouldNotBeNil)
				So(errs, ShouldContainKey, "Email")
			})
		})
	})
}

func TestStructWithMessage(t *testing.T) {
	Convey("Given StructWithMessage", t, func() {
		resetAndInit()

		type User struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
		}

		Convey("When struct has errors and custom messages are provided", func() {
			u := User{}
			errs := validate.StructWithMessage(u, map[string]string{
				"Name": "Silakan isi nama lengkap Anda",
			})

			Convey("Then matching fields should use custom messages", func() {
				So(errs["Name"], ShouldEqual, "Silakan isi nama lengkap Anda")
			})

			Convey("Then non-matching fields keep default messages", func() {
				So(errs["Email"], ShouldNotBeEmpty)
				So(errs["Email"], ShouldNotEqual, "Silakan isi nama lengkap Anda")
			})
		})

		Convey("When struct is valid", func() {
			u := User{Name: "Alice", Email: "alice@example.com"}
			errs := validate.StructWithMessage(u, map[string]string{
				"Name": "custom message",
			})

			Convey("Then it should return nil", func() {
				So(errs, ShouldBeNil)
			})
		})

		Convey("When custom messages map is empty", func() {
			u := User{}
			errs := validate.StructWithMessage(u, map[string]string{})

			Convey("Then it should use default messages", func() {
				So(errs, ShouldNotBeNil)
				So(errs, ShouldContainKey, "Name")
			})
		})
	})
}

func TestFormatErrors(t *testing.T) {
	Convey("Given FormatErrors", t, func() {
		resetAndInit()

		type User struct {
			Name string `validate:"required"`
		}

		Convey("When formatting with a custom translator", func() {
			v := validate.Validator()
			err := v.Struct(User{})
			trans := validate.Translator()

			errs := validate.FormatErrors(err, trans)

			Convey("Then it should return field errors", func() {
				So(errs, ShouldNotBeNil)
				So(errs, ShouldContainKey, "Name")
			})
		})

		Convey("When formatter receives nil translator", func() {
			v := validate.Validator()
			err := v.Struct(User{})
			errs := validate.FormatErrors(err, nil)

			Convey("Then it should use fallback messages", func() {
				So(errs, ShouldNotBeNil)
				So(errs, ShouldContainKey, "Name")
			})
		})
	})
}

// ---------------------------------------------------------------------------
// Custom validations (built-in)
// ---------------------------------------------------------------------------

func TestBuiltinValidations(t *testing.T) {
	Convey("Given the built-in custom validations", t, func() {
		resetAndInit()

		Convey("uuid_any should accept valid UUIDs", func() {
			type T struct {
				ID string `validate:"uuid_any"`
			}
			errs := validate.Struct(T{ID: "550e8400-e29b-41d4-a716-446655440000"})
			So(errs, ShouldBeNil)
		})

		Convey("uuid_any should reject invalid UUIDs", func() {
			type T struct {
				ID string `validate:"uuid_any"`
			}
			errs := validate.Struct(T{ID: "not-a-uuid"})
			So(errs, ShouldNotBeNil)
			So(errs, ShouldContainKey, "ID")
		})

		Convey("safe_text should accept safe characters", func() {
			type T struct {
				Text string `validate:"safe_text"`
			}
			errs := validate.Struct(T{Text: "Hello, World (2026)"})
			So(errs, ShouldBeNil)
		})

		Convey("safe_text should reject dangerous characters", func() {
			type T struct {
				Text string `validate:"safe_text"`
			}
			errs := validate.Struct(T{Text: "<script>alert('xss')</script>"})
			So(errs, ShouldNotBeNil)
		})
	})
}

// ---------------------------------------------------------------------------
// RegisterCustom
// ---------------------------------------------------------------------------

func TestRegisterCustom(t *testing.T) {
	Convey("Given RegisterCustom", t, func() {
		resetAndInit()

		Convey("When registering a custom 'even' validation", func() {
			validate.RegisterCustom("even", func(fl validator.FieldLevel) bool {
				return fl.Field().Int()%2 == 0
			}, "Harus bilangan genap")

			type T struct {
				Num int `validate:"even"`
			}

			Convey("Then it should pass for even numbers", func() {
				errs := validate.Struct(T{Num: 4})
				So(errs, ShouldBeNil)
			})

			Convey("Then it should fail for odd numbers", func() {
				errs := validate.Struct(T{Num: 3})
				So(errs, ShouldNotBeNil)
				So(errs, ShouldContainKey, "Num")
			})
		})
	})
}
