package config

import (
	"fmt"
	"github.com/earcamone/gwy-playground/services/books"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/earcamone/gwy-playground/api/middleware/errorscheme"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("New Default Config", func(t *testing.T) {
		c := New()

		assert.Equal(t, "unknown: check CI/CD workflow", c.Version, "version should be default")
		assert.Equal(t, ":8080", c.Address, "address should be default")
		assert.Equal(t, GracefulTimeout, c.ShutdownTimeout, "shutdown timeout should be default")
		assert.NotNil(t, c.ShutdownFn, "shutdown fn should be set (empty func)")
		assert.Nil(t, c.ErrorResponseFn, "error response fn should be nil by default")
		assert.Empty(t, c.Middlewares, "middlewares should be empty by default")
		assert.NotNil(t, c.Library, "library should be initialized")

		// Test ShutdownFn is callable (won’t do much, just shouldn’t panic)
		assert.NotPanics(t, func() { c.ShutdownFn() }, "default shutdown fn should be safe")
	})

	t.Run("WithVersion", func(t *testing.T) {
		c := New(WithVersion("v1.2.3"))
		assert.Equal(t, "v1.2.3", c.Version, "version should be overridden")
	})

	t.Run("WithAddress", func(t *testing.T) {
		c := New(WithAddress(":9090"))
		assert.Equal(t, ":9090", c.Address, "address should be overridden")
	})

	t.Run("WithGracefulTimeout", func(t *testing.T) {
		c := New(WithGracefulTimeout(5 * time.Second))
		assert.Equal(t, 5*time.Second, c.ShutdownTimeout, "shutdown timeout should be overridden")
	})

	t.Run("WithShutdownFn", func(t *testing.T) {
		called := false
		shutdownFn := func() { called = true }

		c := New(WithShutdownFn(shutdownFn))
		c.ShutdownFn()

		assert.True(t, called, "custom shutdown fn should be called")
	})

	t.Run("WithErrorResponseFunc", func(t *testing.T) {
		called := false
		fn := func(w http.ResponseWriter, err *errorscheme.AppError) {
			called = true
		}

		c := New(WithErrorResponseFunc(fn))
		c.ErrorResponseFn(httptest.NewRecorder(), errorscheme.NewAppError(200, "", fmt.Errorf("error")))

		assert.True(t, called, "function should be set correctly in config")
	})

	t.Run("WithMiddleware", func(t *testing.T) {
		dummyMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			})
		}

		c := New(WithMiddleware(dummyMiddleware))
		assert.Len(t, c.Middlewares, 1, "should add one middleware")
	})

	t.Run("WithDependency Library", func(t *testing.T) {
		mockLib := books.NewLibrary()
		c := New(WithDependency(mockLib))

		assert.Equal(t, mockLib, c.Library, "library should be overridden")
	})

	t.Run("WithDependency Panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for unknown dependency")
			}
		}()

		New(WithDependency("not a valid dependency integrated in WithDependency"))
	})
}
