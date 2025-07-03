package errorscheme

import (
	"context"
	"fmt"
	"net/http"
)

// key private type is used to encapsulate error middleware
// internal key in the request context to avoid clients
// accidentally overwriting its value :P
type key string

// errSetterKey is the key that holds the function that sets
// the exit error in the request "context", called by client
// in the request handlers.
var errSetterKey = key("errSetterKey")

// errSetterFn defines the interface of the function set in the
// request context to allow client, through WithError(), to set
// ahead in the handler the corresponding request error.
type errSetterFn func(*AppError)

// WithError is used by clients in handlers to set errors so the
// centralized error handling middleware can handle them gracefully
func WithError(r *http.Request, err *AppError) {
	r.Context().Value(errSetterKey).(errSetterFn)(err)
}

// ErrResponseFn defines the interface of the function used to
// build the error responses inside the error middleware, by
// default gapi has a built-in one, though clients can replace
// it using WithErrorResponseFunc() to build their own compliant
// error responses. Function receives request writer and error.
type ErrResponseFn func(w http.ResponseWriter, err *AppError)

// ErrorScheme integrates a global Panic handler and a centralized
// errors handling scheme that allow individual route handlers to
// either return their corresponding response or just register in
// the request context an error so this middleware takes care of
// converting it to a response error.
//
// The error scheme still allows client to write its own errors,
// in which case (provided an unhandled panic was not catch), the
// middleware won't override the client written response, allowing
// client to gradually switch its route handlers to this scheme if
// it's integrating it in an existent project.
//
// The scheme used by the middleware to assess if the client already
// wrote a response is wrapping the default http.ResponseWrite with
// the simple responseWriterWrapper struct implemented in this file.
func ErrorScheme(fn ErrResponseFn) func(next http.Handler) http.Handler {
	responseFn := errorResponseHandler

	if fn != nil {
		responseFn = fn
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			//
			// Initialize middleware error tracking scheme in context.
			//
			// We use this approach to avoid breaking context immutability purism,
			// for example updating ahead in the chain a pointer in the handler's
			// context (there is no way for a handler to derive the received context
			// with a new value and overwrite the http.Request one like inside any
			// middleware so "middleware exit chain" can access the new value).
			//
			// So.. we simply set in the handler context a static function that the
			// client can call (through the SetError wrapper) in the event of an error
			// which will simply set the scoped err variable in the context of each
			// received request.
			//

			var err *AppError

			ctx := context.WithValue(r.Context(), errSetterKey, errSetterFn(func(e *AppError) {
				err = e
			}))

			r = r.WithContext(ctx)

			// Panic handler

			defer func() {
				if e := recover(); e != nil {
					appErr := NewAppError(
						http.StatusInternalServerError,
						"Internal server error",
						fmt.Errorf("%v", e),
					)

					responseFn(w, appErr)
				}
			}()

			// Wrap http.ResponseWriter with simple abstraction
			// that will track if client explicitly wrote on
			// response buffer, otherwise giving control to
			// error handling middleware to handle response
			wrapper := newResponseWriterWrapper(w)
			next.ServeHTTP(wrapper, r)

			// Handle error response if client didn't write one
			if err != nil && wrapper.written == false {
				responseFn(w, err)
			}
		})
	}
}
