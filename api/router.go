package api

import (
	"github.com/earcamone/gwy-playground/api/middleware/errorscheme"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/earcamone/gwy-playground/api/config"
	"github.com/earcamone/gwy-playground/api/routes/books"
)

func New(c config.Config) chi.Router {
	// Create New Chi Router
	r := chi.NewRouter()

	r.Use(
		// built-in middlewares
		errorscheme.ErrorScheme(c.ErrorResponseFn),

		// third-party middlewares
		middleware.RealIP,
		middleware.RequestID,
	)

	// Install client supplied middlewares
	r.Use(c.Middlewares...)

	// Register API Routes (located in isolated packages at api/routes)
	r.Mount("/books", books.Books(c))

	return r
}
