package config

import (
	"fmt"
	"net/http"
	"time"

	"github.com/earcamone/gwy-playground/api/middleware/errorscheme"
	"github.com/earcamone/gwy-playground/services/books"
)

type Config struct {
	// Version holds the API version, recommended to set
	// with the CI/CD workflow using release branch version
	Version string

	// Address holds the host and port you want
	// the API to listen incoming connections
	Address string

	// ShutdownTimeout defines the time given
	// to the API to gracefully shutdown
	ShutdownTimeout time.Duration

	// ShutdownFn is a function the API
	// will call when a shutdown signal is received
	ShutdownFn func()

	// ErrorResponseFn holds the function used
	// to build requests error responses when
	// client uses WithError(), relying in gapi
	// centralized error handling middleware.
	ErrorResponseFn errorscheme.ErrResponseFn

	// Middlewares holds additional Middlewares
	// client might want to add to API router
	Middlewares []func(next http.Handler) http.Handler

	// Library holds out bogus books library service
	Library books.Library
}

type WithFunc func(Config) Config

func New(fn ...WithFunc) Config {
	c := Config{
		Version: version,
		Address: ":8080",

		ShutdownFn:      func() {},
		ShutdownTimeout: GracefulTimeout,

		// API Services Dependencies
		Library: books.NewLibrary(),
	}

	for _, f := range fn {
		c = f(c)
	}

	return c
}

func WithVersion(v string) func(Config) Config {
	return func(config Config) Config {
		config.Version = v
		return config
	}
}

func WithAddress(addr string) func(Config) Config {
	return func(config Config) Config {
		config.Address = addr
		return config
	}
}

func WithErrorResponseFunc(fn errorscheme.ErrResponseFn) func(Config) Config {
	return func(config Config) Config {
		config.ErrorResponseFn = fn
		return config
	}
}

func WithShutdownFn(fn func()) func(Config) Config {
	return func(config Config) Config {
		config.ShutdownFn = fn
		return config
	}
}

func WithGracefulTimeout(t time.Duration) func(Config) Config {
	return func(config Config) Config {
		config.ShutdownTimeout = t
		return config
	}
}

func WithMiddleware(m func(next http.Handler) http.Handler) WithFunc {
	return func(config Config) Config {
		config.Middlewares = append(config.Middlewares, m)
		return config
	}
}

func WithDependency(dep any) WithFunc {
	return func(config Config) Config {
		switch d := dep.(type) {
		case books.Library:
			config.Library = d

		default:
			panic(fmt.Sprintf("injected dependency not supported, please ensure you updated the WithDependency config function correctly: %v", d))
		}

		return config
	}
}
