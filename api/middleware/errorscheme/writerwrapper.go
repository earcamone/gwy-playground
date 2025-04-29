package errorscheme

import "net/http"

// responseWriterWrapper is a simple http.ResponseWriter that holds a flag
// to acknowledge if a response was written from within a route handler,
// error handling middleware will take control of error responses only if
// client didn't write response buffer (taking ownership of error handling)
type responseWriterWrapper struct {
	http.ResponseWriter
	written bool
}

func (r *responseWriterWrapper) Write(b []byte) (int, error) {
	r.written = true
	return r.ResponseWriter.Write(b)
}

func (r *responseWriterWrapper) WriteHeader(code int) {
	r.written = true
	r.ResponseWriter.WriteHeader(code)
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{w, false}
}
