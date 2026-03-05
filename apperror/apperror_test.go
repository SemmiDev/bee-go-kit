package apperror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/apperror"
)

// ---------------------------------------------------------------------------
// Error struct methods
// ---------------------------------------------------------------------------

func TestError_Error(t *testing.T) {
	Convey("Given an apperror.Error", t, func() {
		Convey("When there is no wrapped error", func() {
			err := apperror.NotFound("user not found")

			Convey("Then Error() should return code and message", func() {
				So(err.Error(), ShouldEqual, "[ERR_NOT_FOUND] user not found")
			})
		})

		Convey("When there is a wrapped error", func() {
			cause := fmt.Errorf("db connection refused")
			err := apperror.Internal("database error", cause)

			Convey("Then Error() should include the underlying error", func() {
				So(err.Error(), ShouldContainSubstring, "database error")
				So(err.Error(), ShouldContainSubstring, "db connection refused")
			})
		})
	})
}

func TestError_Unwrap(t *testing.T) {
	Convey("Given an apperror.Error wrapping another error", t, func() {
		cause := fmt.Errorf("root cause")
		err := apperror.Internal("failed", cause)

		Convey("When Unwrap is called", func() {
			unwrapped := err.Unwrap()

			Convey("Then it should return the underlying error", func() {
				So(unwrapped, ShouldEqual, cause)
			})
		})

		Convey("When errors.Is is used", func() {
			Convey("Then it should find the wrapped error in the chain", func() {
				So(errors.Is(err, cause), ShouldBeTrue)
			})
		})
	})
}

func TestError_HTTPStatus(t *testing.T) {
	Convey("Given various apperror types", t, func() {
		Convey("Then HTTPStatus should return the correct code", func() {
			So(apperror.NotFound("x").HTTPStatus(), ShouldEqual, http.StatusNotFound)
			So(apperror.BadRequest("x").HTTPStatus(), ShouldEqual, http.StatusBadRequest)
			So(apperror.Unauthorized("x").HTTPStatus(), ShouldEqual, http.StatusUnauthorized)
			So(apperror.Forbidden("x").HTTPStatus(), ShouldEqual, http.StatusForbidden)
			So(apperror.Conflict("x").HTTPStatus(), ShouldEqual, http.StatusConflict)
			So(apperror.Internal("x", nil).HTTPStatus(), ShouldEqual, http.StatusInternalServerError)
			So(apperror.ServiceUnavailable("x").HTTPStatus(), ShouldEqual, http.StatusServiceUnavailable)
			So(apperror.Unprocessable("x").HTTPStatus(), ShouldEqual, http.StatusUnprocessableEntity)
		})
	})
}

func TestError_Wrap(t *testing.T) {
	Convey("Given an existing apperror", t, func() {
		original := apperror.NotFound("user not found")
		newCause := fmt.Errorf("sql: no rows")

		Convey("When Wrap is called with a new cause", func() {
			wrapped := original.Wrap(newCause)

			Convey("Then it should preserve code, error code, and message", func() {
				So(wrapped.Code, ShouldEqual, original.Code)
				So(wrapped.ErrorCode, ShouldEqual, original.ErrorCode)
				So(wrapped.Message, ShouldEqual, original.Message)
			})

			Convey("Then it should replace the underlying error", func() {
				So(wrapped.Unwrap(), ShouldEqual, newCause)
			})

			Convey("Then it should not mutate the original", func() {
				So(original.Unwrap(), ShouldBeNil)
			})
		})
	})
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

func TestConstructors(t *testing.T) {
	Convey("Given the apperror constructors", t, func() {
		Convey("NotFound should create a 404 error", func() {
			err := apperror.NotFound("gone")
			So(err.Code, ShouldEqual, http.StatusNotFound)
			So(err.ErrorCode, ShouldEqual, apperror.ErrCodeNotFound)
			So(err.Message, ShouldEqual, "gone")
		})

		Convey("NotFoundf should format the message", func() {
			err := apperror.NotFoundf("user %d not found", 42)
			So(err.Message, ShouldEqual, "user 42 not found")
			So(err.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("BadRequest should create a 400 error", func() {
			err := apperror.BadRequest("bad input")
			So(err.Code, ShouldEqual, http.StatusBadRequest)
			So(err.ErrorCode, ShouldEqual, apperror.ErrCodeBadRequest)
		})

		Convey("BadRequestWrap should wrap an underlying error", func() {
			cause := fmt.Errorf("parse error")
			err := apperror.BadRequestWrap("invalid JSON", cause)
			So(err.Code, ShouldEqual, http.StatusBadRequest)
			So(err.Unwrap(), ShouldEqual, cause)
		})

		Convey("Validation should create a 400 with validation code", func() {
			err := apperror.Validation("field required")
			So(err.Code, ShouldEqual, http.StatusBadRequest)
			So(err.ErrorCode, ShouldEqual, apperror.ErrCodeValidation)
		})

		Convey("Duplicate should format resource/field/value", func() {
			err := apperror.Duplicate("User", "email", "a@b.com")
			So(err.Code, ShouldEqual, http.StatusConflict)
			So(err.ErrorCode, ShouldEqual, apperror.ErrCodeDuplicate)
			So(err.Message, ShouldContainSubstring, "User")
			So(err.Message, ShouldContainSubstring, "email")
			So(err.Message, ShouldContainSubstring, "a@b.com")
		})

		Convey("InternalFromErr should use err as both message source and cause", func() {
			cause := fmt.Errorf("timeout")
			err := apperror.InternalFromErr(cause)
			So(err.Code, ShouldEqual, http.StatusInternalServerError)
			So(err.Unwrap(), ShouldEqual, cause)
		})
	})
}

// ---------------------------------------------------------------------------
// Generic constructors (extensibility)
// ---------------------------------------------------------------------------

func TestGenericConstructors(t *testing.T) {
	Convey("Given the generic constructors", t, func() {
		Convey("New should create an error with any status/code", func() {
			err := apperror.New(http.StatusPaymentRequired, "ERR_PAYMENT", "payment required")
			So(err.Code, ShouldEqual, http.StatusPaymentRequired)
			So(err.ErrorCode, ShouldEqual, "ERR_PAYMENT")
			So(err.Message, ShouldEqual, "payment required")
			So(err.Unwrap(), ShouldBeNil)
		})

		Convey("Newf should format the message", func() {
			err := apperror.Newf(http.StatusTooManyRequests, "ERR_RATE_LIMIT", "rate limited: %d req/s", 100)
			So(err.Code, ShouldEqual, http.StatusTooManyRequests)
			So(err.Message, ShouldEqual, "rate limited: 100 req/s")
		})

		Convey("NewWithErr should wrap an underlying error", func() {
			cause := fmt.Errorf("upstream timeout")
			err := apperror.NewWithErr(http.StatusBadGateway, "ERR_UPSTREAM", "upstream failed", cause)
			So(err.Code, ShouldEqual, http.StatusBadGateway)
			So(err.ErrorCode, ShouldEqual, "ERR_UPSTREAM")
			So(err.Unwrap(), ShouldEqual, cause)
		})
	})
}

// ---------------------------------------------------------------------------
// Checkers
// ---------------------------------------------------------------------------

func TestCheckers(t *testing.T) {
	Convey("Given various errors", t, func() {
		notFoundErr := apperror.NotFound("x")
		badReqErr := apperror.BadRequest("x")
		unauthErr := apperror.Unauthorized("x")
		forbiddenErr := apperror.Forbidden("x")
		conflictErr := apperror.Conflict("x")
		internalErr := apperror.Internal("x", nil)
		plainErr := fmt.Errorf("just a plain error")

		Convey("IsNotFound should match only 404 errors", func() {
			So(apperror.IsNotFound(notFoundErr), ShouldBeTrue)
			So(apperror.IsNotFound(badReqErr), ShouldBeFalse)
			So(apperror.IsNotFound(plainErr), ShouldBeFalse)
		})

		Convey("IsBadRequest should match only 400 errors", func() {
			So(apperror.IsBadRequest(badReqErr), ShouldBeTrue)
			So(apperror.IsBadRequest(notFoundErr), ShouldBeFalse)
		})

		Convey("IsUnauthorized should match only 401 errors", func() {
			So(apperror.IsUnauthorized(unauthErr), ShouldBeTrue)
			So(apperror.IsUnauthorized(notFoundErr), ShouldBeFalse)
		})

		Convey("IsForbidden should match only 403 errors", func() {
			So(apperror.IsForbidden(forbiddenErr), ShouldBeTrue)
			So(apperror.IsForbidden(notFoundErr), ShouldBeFalse)
		})

		Convey("IsConflict should match only 409 errors", func() {
			So(apperror.IsConflict(conflictErr), ShouldBeTrue)
			So(apperror.IsConflict(notFoundErr), ShouldBeFalse)
		})

		Convey("IsInternal should match only 500 errors", func() {
			So(apperror.IsInternal(internalErr), ShouldBeTrue)
			So(apperror.IsInternal(notFoundErr), ShouldBeFalse)
		})
	})
}

// ---------------------------------------------------------------------------
// Utilities
// ---------------------------------------------------------------------------

func TestHTTPStatus(t *testing.T) {
	Convey("Given the HTTPStatus utility", t, func() {
		Convey("When the error is an apperror", func() {
			Convey("Then it should return the error's status code", func() {
				So(apperror.HTTPStatus(apperror.NotFound("x")), ShouldEqual, 404)
				So(apperror.HTTPStatus(apperror.Forbidden("x")), ShouldEqual, 403)
			})
		})

		Convey("When the error is a plain error", func() {
			Convey("Then it should default to 500", func() {
				So(apperror.HTTPStatus(fmt.Errorf("boom")), ShouldEqual, 500)
			})
		})
	})
}

func TestMessage(t *testing.T) {
	Convey("Given the Message utility", t, func() {
		Convey("When the error is an apperror", func() {
			err := apperror.NotFound("user not found")
			So(apperror.Message(err), ShouldEqual, "user not found")
		})

		Convey("When the error is a plain error", func() {
			err := fmt.Errorf("something broke")
			So(apperror.Message(err), ShouldEqual, "something broke")
		})
	})
}

func TestCode(t *testing.T) {
	Convey("Given the Code utility", t, func() {
		Convey("When the error is an apperror", func() {
			err := apperror.NotFound("x")
			So(apperror.Code(err), ShouldEqual, apperror.ErrCodeNotFound)
		})

		Convey("When the error is a plain error", func() {
			err := fmt.Errorf("x")
			So(apperror.Code(err), ShouldBeEmpty)
		})
	})
}
