package errorscheme

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	t.Run("Error Method", func(t *testing.T) {
		err := &AppError{
			Code:     400,
			Message:  "Bad Request, bro",
			Internal: errors.New("oops"),
		}

		assert.Equal(t, "Bad Request, bro", err.Error(), "Error() should return Message")
	})

	t.Run("NewAppError", func(t *testing.T) {
		internalErr := errors.New("internal glitch")
		appErr := NewAppError(500, "Server borked", internalErr)

		assert.Equal(t, 500, appErr.Code, "Code should be set")
		assert.Equal(t, "Server borked", appErr.Message, "Message should be set")
		assert.Equal(t, internalErr, appErr.Internal, "Internal should be set")
	})
}
