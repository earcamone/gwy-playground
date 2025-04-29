package errorscheme

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorScheme(t *testing.T) {
	t.Run("WithError Sets Error", func(t *testing.T) {
		var capturedErr *AppError
		setter := errSetterFn(func(e *AppError) { capturedErr = e })
		ctx := context.WithValue(context.Background(), errSetterKey, setter)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)

		err := AppError{Code: 404, Message: "Lost, bro"}
		WithError(req, &err)

		assert.Equal(t, 404, capturedErr.Code, "should set error code")
		assert.Equal(t, "Lost, bro", capturedErr.Message, "should set error message")
	})

	t.Run("WithError No Middleware", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		err := AppError{Code: 500, Message: "No setter, bro"}

		// Should panic gracefully (no setter in context)
		assert.Panics(t, func() { WithError(req, &err) }, "should panic without setter")
	})

	t.Run("ErrorScheme No Error No Panic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Chillin, bro"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		ErrorScheme(nil)(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "should return 200 OK")
		assert.Equal(t, "Chillin, bro", w.Body.String(), "should pass through handler response")
	})

	t.Run("ErrorScheme Panic Caught", func(t *testing.T) {
		// custom panic handler to mimic panic inside handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("Freakout, bro!")
		})

		// call error middleware with panic handler

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		ErrorScheme(nil)(handler).ServeHTTP(w, req)

		// read middleware response after handler triggered panic

		var resp errorResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		// ensure built-in panic error handling response is returned
		assert.Equal(t, http.StatusInternalServerError, w.Code, "should return 500 on panic")
		assert.Equal(t, 500, resp.Status, "response status should be 500")
		assert.Equal(t, "Internal server error", resp.Message, "response message should be set")
	})

	t.Run("ErrorScheme Error Set No Client Write", func(t *testing.T) {
		// simple handler that mimics client setting error in handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			WithError(r, &AppError{Code: 400, Message: "Bad move, bro"})
		})

		// call our middleware with handler setting error wrapping it

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		ErrorScheme(nil)(handler).ServeHTTP(w, req)

		// read middleware response after handler set error

		var resp errorResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		// ensure built-in error handling response is returned
		assert.Equal(t, 400, w.Code, "should return error code")
		assert.Equal(t, 400, resp.Status, "response status should match error")
		assert.Equal(t, "Bad move, bro", resp.Message, "response message should match error")
	})

	t.Run("ErrorScheme Error Set Client Wrote", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			WithError(r, &AppError{Code: 400, Message: "Bad move, bro"})

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Client rules, bro"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		ErrorScheme(nil)(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "should respect client status")
		assert.Equal(t, "Client rules, bro", w.Body.String(), "should respect client response")
	})

	t.Run("ErrorScheme Custom ErrResponseFn", func(t *testing.T) {
		customFn := func(w http.ResponseWriter, err *AppError) {
			w.WriteHeader(err.Code)
			w.Write([]byte("Custom error, bro"))
		}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			WithError(r, &AppError{Code: 418, Message: "Teapot, bro"})
		})
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		ErrorScheme(customFn)(handler).ServeHTTP(w, req)

		assert.Equal(t, 418, w.Code, "should use custom error code")
		assert.Equal(t, "Custom error, bro", w.Body.String(), "should use custom response")
	})
}
