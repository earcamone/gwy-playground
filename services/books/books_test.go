package books

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLibraryBasicOperations(t *testing.T) {
	lib := NewLibrary()

	// Test Add
	book1 := &Book{Id: "1", Name: "Go Programming", Pages: 300}
	err := lib.Add(book1)
	assert.NoError(t, err, "Adding first book should succeed")

	// Test Add duplicate
	err = lib.Add(book1)
	assert.Error(t, err, "Adding duplicate book should fail")
	assert.Equal(t, "book already exists", err.Error(), "Expected duplicate error message")

	// Test Get
	book, err := lib.Get("1")
	assert.NoError(t, err, "Getting existing book should succeed")
	assert.Equal(t, book1, book, "Retrieved book should match added book")

	// Test Get non-existent
	_, err = lib.Get("999")
	assert.Error(t, err, "Getting non-existent book should fail")
	assert.Equal(t, "book not found", err.Error(), "Expected not found error")

	// Test Remove
	err = lib.Remove("1")
	assert.NoError(t, err, "Removing existing book should succeed")

	// Test Remove non-existent
	err = lib.Remove("1")
	assert.Error(t, err, "Removing non-existent book should fail")
	assert.Equal(t, "book not found", err.Error(), "Expected not found error")

	// Test Get non-existent
	_, err = lib.Get("1")
	assert.Error(t, err, "Getting removing book should fail")
	assert.Equal(t, "book not found", err.Error(), "Expected not found error")
}
