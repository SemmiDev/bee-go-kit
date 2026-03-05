package httputil_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/semmidev/bee-go-kit/apperror"
	"github.com/semmidev/bee-go-kit/httputil"
	"github.com/semmidev/bee-go-kit/pagination"
)

// ---------------------------------------------------------------------------
// Response constructors
// ---------------------------------------------------------------------------

func TestSuccess(t *testing.T) {
	Convey("Given Success", t, func() {
		resp := httputil.Success("ok", map[string]string{"id": "1"})

		Convey("Then it should create a success response", func() {
			So(resp.Success, ShouldBeTrue)
			So(resp.Message, ShouldEqual, "ok")
			So(resp.Data, ShouldNotBeNil)
			So(resp.ErrorCode, ShouldBeEmpty)
		})
	})
}

func TestSuccessWithPaging(t *testing.T) {
	Convey("Given SuccessWithPaging", t, func() {
		paging := &pagination.Paging{CurrentPage: 1, PerPage: 10, TotalData: 100}
		resp := httputil.SuccessWithPaging("list ok", []string{"a"}, paging)

		Convey("Then it should include paging metadata", func() {
			So(resp.Success, ShouldBeTrue)
			So(resp.Paging, ShouldNotBeNil)
			So(resp.Paging.TotalData, ShouldEqual, 100)
		})
	})
}

func TestValidationError(t *testing.T) {
	Convey("Given ValidationError", t, func() {
		errs := map[string]string{"email": "required"}
		resp := httputil.ValidationError("validation failed", errs)

		Convey("Then it should have validation errors", func() {
			So(resp.Success, ShouldBeFalse)
			So(resp.ErrorCode, ShouldEqual, apperror.ErrCodeValidationFailed)
			So(resp.Validation, ShouldContainKey, "email")
		})
	})
}

func TestFromError(t *testing.T) {
	Convey("Given FromError", t, func() {
		Convey("When error is an apperror", func() {
			err := apperror.NotFound("user not found")
			status, resp := httputil.FromError(err)

			So(status, ShouldEqual, http.StatusNotFound)
			So(resp.Success, ShouldBeFalse)
			So(resp.ErrorCode, ShouldEqual, apperror.ErrCodeNotFound)
			So(resp.Message, ShouldEqual, "user not found")
		})

		Convey("When error is a plain error", func() {
			err := fmt.Errorf("something broke")
			status, resp := httputil.FromError(err)

			So(status, ShouldEqual, http.StatusInternalServerError)
			So(resp.Success, ShouldBeFalse)
			So(resp.ErrorCode, ShouldEqual, apperror.ErrCodeInternal)
		})
	})
}

// ---------------------------------------------------------------------------
// ErrorTransformer (extensibility)
// ---------------------------------------------------------------------------

type customError struct {
	Status  int
	ErrCode string
	Msg     string
}

func (e *customError) Error() string { return e.Msg }

func TestErrorTransformer(t *testing.T) {
	Convey("Given SetErrorTransformer", t, func() {
		// Clean up after test
		Reset(func() {
			httputil.SetErrorTransformer(nil)
		})

		Convey("When a custom transformer handles the error", func() {
			httputil.SetErrorTransformer(func(err error) (int, httputil.Response) {
				if ce, ok := err.(*customError); ok {
					return ce.Status, httputil.Response{
						Success:   false,
						ErrorCode: ce.ErrCode,
						Message:   ce.Msg,
					}
				}
				return 0, httputil.Response{} // fall through
			})

			status, resp := httputil.FromError(&customError{
				Status:  http.StatusPaymentRequired,
				ErrCode: "ERR_PAYMENT",
				Msg:     "payment required",
			})

			Convey("Then it should use the custom transformer result", func() {
				So(status, ShouldEqual, http.StatusPaymentRequired)
				So(resp.ErrorCode, ShouldEqual, "ERR_PAYMENT")
				So(resp.Message, ShouldEqual, "payment required")
			})
		})

		Convey("When the custom transformer returns 0 (falls through)", func() {
			httputil.SetErrorTransformer(func(err error) (int, httputil.Response) {
				return 0, httputil.Response{}
			})

			status, resp := httputil.FromError(apperror.NotFound("not found"))

			Convey("Then it should use the default apperror logic", func() {
				So(status, ShouldEqual, http.StatusNotFound)
				So(resp.ErrorCode, ShouldEqual, apperror.ErrCodeNotFound)
			})
		})

		Convey("When no transformer is set", func() {
			httputil.SetErrorTransformer(nil)

			status, _ := httputil.FromError(fmt.Errorf("plain error"))

			Convey("Then it should use default 500 logic", func() {
				So(status, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

// ---------------------------------------------------------------------------
// JSON writers
// ---------------------------------------------------------------------------

func TestWriteJSON(t *testing.T) {
	Convey("Given WriteJSON", t, func() {
		Convey("When writing a valid body", func() {
			w := httptest.NewRecorder()
			httputil.WriteJSON(w, http.StatusOK, map[string]string{"key": "value"})

			Convey("Then it should set correct headers and status", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Header().Get("Content-Type"), ShouldContainSubstring, "application/json")
			})

			Convey("Then it should encode as JSON", func() {
				var body map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &body)
				So(err, ShouldBeNil)
				So(body["key"], ShouldEqual, "value")
			})
		})
	})
}

func TestWriteSuccess(t *testing.T) {
	Convey("Given WriteSuccess", t, func() {
		w := httptest.NewRecorder()
		httputil.WriteSuccess(w, http.StatusOK, "done", nil)

		Convey("Then it should return 200 with success=true", func() {
			So(w.Code, ShouldEqual, http.StatusOK)

			var resp httputil.Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp.Success, ShouldBeTrue)
			So(resp.Message, ShouldEqual, "done")
		})
	})
}

func TestWriteError(t *testing.T) {
	Convey("Given WriteError", t, func() {
		Convey("When error is an apperror", func() {
			w := httptest.NewRecorder()
			httputil.WriteError(w, apperror.Forbidden("denied"))

			So(w.Code, ShouldEqual, http.StatusForbidden)

			var resp httputil.Response
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			So(resp.Success, ShouldBeFalse)
			So(resp.ErrorCode, ShouldEqual, apperror.ErrCodeForbidden)
		})
	})
}

func TestWriteValidationError(t *testing.T) {
	Convey("Given WriteValidationError", t, func() {
		w := httptest.NewRecorder()
		httputil.WriteValidationError(w, "invalid input", map[string]string{"name": "required"})

		Convey("Then it should return 400 with validation errors", func() {
			So(w.Code, ShouldEqual, http.StatusBadRequest)

			var resp httputil.Response
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			So(resp.Validation["name"], ShouldEqual, "required")
		})
	})
}

// ---------------------------------------------------------------------------
// Error handlers
// ---------------------------------------------------------------------------

func TestErrorHandlers(t *testing.T) {
	Convey("Given ErrorHandlers with default config", t, func() {
		handlers := httputil.NewErrorHandlers(httputil.DefaultHandlerConfig())

		tests := []struct {
			name       string
			handler    func(http.ResponseWriter, *http.Request)
			wantStatus int
		}{
			{"NotFound", handlers.NotFound, http.StatusNotFound},
			{"MethodNotAllowed", handlers.MethodNotAllowed, http.StatusMethodNotAllowed},
			{"InternalError", handlers.InternalError, http.StatusInternalServerError},
			{"BadRequest", handlers.BadRequest, http.StatusBadRequest},
			{"Unauthorized", handlers.Unauthorized, http.StatusUnauthorized},
			{"Forbidden", handlers.Forbidden, http.StatusForbidden},
		}

		for _, tc := range tests {
			Convey(fmt.Sprintf("%s should return %d", tc.name, tc.wantStatus), func() {
				w := httptest.NewRecorder()
				tc.handler(w, httptest.NewRequest("GET", "/", nil))
				So(w.Code, ShouldEqual, tc.wantStatus)

				var resp httputil.Response
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
				So(resp.Success, ShouldBeFalse)
			})
		}
	})
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

func TestRecoveryMiddleware(t *testing.T) {
	Convey("Given the Recovery middleware", t, func() {
		Convey("When the handler panics", func() {
			panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("unexpected error!")
			})
			handler := httputil.Recovery(panicHandler)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			handler.ServeHTTP(w, r)

			Convey("Then it should recover and return 500 JSON", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)

				var resp httputil.Response
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
				So(resp.Success, ShouldBeFalse)
			})
		})

		Convey("When the handler does not panic", func() {
			okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handler := httputil.Recovery(okHandler)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			handler.ServeHTTP(w, r)

			Convey("Then it should pass through normally", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	Convey("Given the RequestID middleware", t, func() {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := httputil.RequestID(inner)

		Convey("When no X-Request-Id is provided", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			handler.ServeHTTP(w, r)

			Convey("Then it should generate a new one", func() {
				id := w.Header().Get("X-Request-Id")
				So(id, ShouldNotBeEmpty)
			})
		})

		Convey("When X-Request-Id is already provided", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Set("X-Request-Id", "existing-id-123")
			handler.ServeHTTP(w, r)

			Convey("Then it should preserve the existing one", func() {
				So(w.Header().Get("X-Request-Id"), ShouldEqual, "existing-id-123")
			})
		})
	})
}

func TestGetRequestID(t *testing.T) {
	Convey("Given GetRequestID", t, func() {
		Convey("When request has X-Request-Id", func() {
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Set("X-Request-Id", "test-id")
			So(httputil.GetRequestID(r), ShouldEqual, "test-id")
		})

		Convey("When request has no X-Request-Id", func() {
			r := httptest.NewRequest("GET", "/", nil)
			So(httputil.GetRequestID(r), ShouldBeEmpty)
		})
	})
}

func TestNoCacheMiddleware(t *testing.T) {
	Convey("Given the NoCache middleware", t, func() {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := httputil.NoCache(inner)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		handler.ServeHTTP(w, r)

		Convey("Then it should set no-cache headers", func() {
			So(w.Header().Get("Cache-Control"), ShouldContainSubstring, "no-cache")
			So(w.Header().Get("Pragma"), ShouldEqual, "no-cache")
			So(w.Header().Get("Expires"), ShouldEqual, "0")
		})
	})
}
