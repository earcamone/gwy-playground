package errorscheme

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorResponse(t *testing.T) {
	t.Run("errorResponseHandler Basic", func(t *testing.T) {
		w := httptest.NewRecorder()
		appErr := NewAppError(403, "No entry, bro", nil)
		errorResponseHandler(w, appErr)

		assert.Equal(t, 403, w.Code, "should set status code")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "should set JSON content type")
		var resp errorResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, 403, resp.Status, "response status should match")
		assert.Equal(t, "No entry, bro", resp.Message, "response message should match")
		assert.Empty(t, resp.Details, "details should be empty without internal error")
	})

	t.Run("errorResponseHandler With Internal Dev", func(t *testing.T) {
		os.Setenv("ENV", "develop")
		defer os.Unsetenv("ENV")

		w := httptest.NewRecorder()
		appErr := NewAppError(500, "Boom, bro", errors.New("kaboom"))
		errorResponseHandler(w, appErr)

		var resp errorResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, 500, resp.Status, "response status should match")
		assert.Equal(t, "Boom, bro", resp.Message, "response message should match")
		assert.Equal(t, "kaboom", resp.Details, "details should include internal error in dev")
	})
}
