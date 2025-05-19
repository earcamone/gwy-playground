package errorscheme

import (
	"net/http/httptest"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestWriterWrapper(t *testing.T) {
	t.Run("responseWriterWrapper Write", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapper := newResponseWriterWrapper(w)
		assert.False(t, wrapper.written, "written should start false")

		body := "Test, bro"
		n, err := wrapper.Write([]byte(body))

		assert.NoError(t, err, "write should succeed")
		assert.Equal(t, len(body), n, "should write correct bytes")
		assert.True(t, wrapper.written, "written should be true after Write")
		assert.Equal(t, "Test, bro", w.Body.String(), "should delegate to underlying writer")
	})

	t.Run("responseWriterWrapper WriteHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapper := newResponseWriterWrapper(w)
		assert.False(t, wrapper.written, "written should start false")

		wrapper.WriteHeader(201)
		assert.True(t, wrapper.written, "written should be true after WriteHeader")
		assert.Equal(t, 201, w.Code, "should delegate status to underlying writer")
	})
}
