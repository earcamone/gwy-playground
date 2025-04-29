package books

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/earcamone/gapi/api/config"
	"github.com/earcamone/gapi/api/middleware/errorscheme"
	"github.com/earcamone/gapi/services/books"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type MockLibrary struct {
	getFn    func(string) (*books.Book, error)
	addFn    func(*books.Book) error
	removeFn func(string) error
}

func (m *MockLibrary) Get(id string) (*books.Book, error) { return m.getFn(id) }
func (m *MockLibrary) Add(book *books.Book) error         { return m.addFn(book) }
func (m *MockLibrary) Remove(id string) error             { return m.removeFn(id) }

func setupRouter(c config.Config, errFn errorscheme.ErrResponseFn) http.Handler {
	r := chi.NewRouter()

	r.Use(errorscheme.ErrorScheme(errFn))
	r.Mount("/", Books(c))
	
	return r
}

func TestBooksSuccess(t *testing.T) {
	lib := books.NewLibrary()
	c := config.Config{Library: lib}
	router := setupRouter(c, nil)

	t.Run("AddBook Success", func(t *testing.T) {
		book := books.Book{Id: "1", Name: "Test Book", Pages: 100}
		body, _ := json.Marshal(book)
		req, _ := http.NewRequest("POST", "/", bytes.NewReader(body)) // Changed to "/"
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "/1", w.Header().Get("Location")) // Now "/1" since mounted at "/"
		var respBook books.Book
		json.Unmarshal(w.Body.Bytes(), &respBook)
		assert.Equal(t, book, respBook)
	})

	t.Run("GetBook Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/1", nil) // Changed to "/1"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respBook books.Book
		json.Unmarshal(w.Body.Bytes(), &respBook)
		assert.Equal(t, books.Book{Id: "1", Name: "Test Book", Pages: 100}, respBook)
	})

	t.Run("RemoveBook Success", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/1", nil) // Changed to "/1"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
	})
}

func TestBooksErrors(t *testing.T) {
	var capturedErr *errorscheme.AppError
	errFn := func(w http.ResponseWriter, err *errorscheme.AppError) {
		capturedErr = err
		w.WriteHeader(err.Code)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Message})
	}

	t.Run("AddBook Invalid Body", func(t *testing.T) {
		c := config.Config{Library: books.NewLibrary()}
		router := setupRouter(c, errFn)
		req, _ := http.NewRequest("POST", "/", bytes.NewReader([]byte("invalid"))) // Changed to "/"
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, http.StatusBadRequest, capturedErr.Code)
		assert.Equal(t, "invalid request body", capturedErr.Message)
	})

	t.Run("AddBook Library Failure", func(t *testing.T) {
		mockLib := &MockLibrary{
			addFn: func(*books.Book) error { return errors.New("mock add error") },
		}
		c := config.Config{Library: mockLib}
		router := setupRouter(c, errFn)
		book := books.Book{Id: "2", Name: "Fail Book", Pages: 200}
		body, _ := json.Marshal(book)
		req, _ := http.NewRequest("POST", "/", bytes.NewReader(body)) // Changed to "/"
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, http.StatusInternalServerError, capturedErr.Code)
		assert.Equal(t, "error adding book", capturedErr.Message)
	})

	t.Run("GetBook Not Found", func(t *testing.T) {
		mockLib := &MockLibrary{
			getFn: func(string) (*books.Book, error) { return nil, errors.New("not found") },
		}
		c := config.Config{Library: mockLib}
		router := setupRouter(c, errFn)
		req, _ := http.NewRequest("GET", "/999", nil) // Changed to "/999"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, http.StatusNotFound, capturedErr.Code)
		assert.Equal(t, "book not found", capturedErr.Message)
	})

	t.Run("RemoveBook Not Found", func(t *testing.T) {
		mockLib := &MockLibrary{
			removeFn: func(string) error { return errors.New("not found") },
		}
		c := config.Config{Library: mockLib}
		router := setupRouter(c, errFn)
		req, _ := http.NewRequest("DELETE", "/999", nil) // Changed to "/999"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, http.StatusNotFound, capturedErr.Code)
		assert.Equal(t, "book not found", capturedErr.Message)
	})
}
