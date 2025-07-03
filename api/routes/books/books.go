package books

import (
	"encoding/json"
	"github.com/earcamone/gwy-playground/api/middleware/errorscheme"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"

	"github.com/earcamone/gwy-playground/api/config"
	"github.com/earcamone/gwy-playground/services/books"
)

func Books(c config.Config) chi.Router {
	r := chi.NewRouter()

	r.Post("/", AddBook(c))
	r.Get("/{id}", GetBook(c))
	r.Delete("/{id}", RemoveBook(c))

	return r
}

func GetBook(c config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get book from library
		id := chi.URLParam(r, "id")
		book, err := c.Library.Get(id)

		if err != nil {
			errorscheme.WithError(r,
				errorscheme.NewAppError(http.StatusNotFound, "book not found", err),
			)
			return
		}

		// Encode book as JSON response
		err = json.NewEncoder(w).Encode(book)

		if err != nil {
			errorscheme.WithError(r,
				errorscheme.NewAppError(http.StatusInternalServerError, "error encoding book", err),
			)
		}
	}
}

func RemoveBook(c config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Remove book from library
		id := chi.URLParam(r, "id")
		err := c.Library.Remove(id)

		if err != nil {
			errorscheme.WithError(r,
				errorscheme.NewAppError(http.StatusNotFound, "book not found", err),
			)
			return
		}

		// Return 204 No Content on successful deletion
		w.WriteHeader(http.StatusNoContent)
	}
}

func AddBook(c config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse received book information
		var book books.Book

		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			errorscheme.WithError(r,
				errorscheme.NewAppError(http.StatusBadRequest, "invalid request body", err),
			)

			return
		}

		// Add book to library
		if err := c.Library.Add(&book); err != nil {
			errorscheme.WithError(r,
				errorscheme.NewAppError(http.StatusInternalServerError, "error adding book", err),
			)

			return
		}

		// Write book successful create response
		w.WriteHeader(http.StatusCreated)

		location := path.Join(r.URL.Path, book.Id)
		w.Header().Set("Location", location) // Use book.Id instead of r.RequestURI

		err := json.NewEncoder(w).Encode(book)

		if err != nil {
			errorscheme.WithError(r,
				errorscheme.NewAppError(http.StatusInternalServerError, "error encoding book", err),
			)

			return
		}
	}
}
