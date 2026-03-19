package main

import (
	"fmt"
	"net/http"
	"time"
)

// Step 5: Middleware — functions that wrap HTTP handlers
// A middleware takes an http.Handler, does something, then calls the next handler
// Pattern: func(next http.Handler) http.Handler

// Logger middleware logs every request with method, path, status code and duration
// Example output: GET /users 200 1.234ms
func Logger(next http.Handler) http.Handler {
	// http.HandlerFunc is an adapter — it lets us use a function as http.Handler
	// This is a common Go pattern: returning an interface by wrapping a function
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record the start time BEFORE the request is handled
		start := time.Now()

		// Wrap the ResponseWriter to capture the status code
		// We need this because the default ResponseWriter doesn't expose the status code
		wrapped := &statusRecorder{ResponseWriter: w, statusCode: 200}

		// Call the next handler in the chain (the actual route handler)
		next.ServeHTTP(wrapped, r)

		// After the handler finishes, calculate duration and log
		duration := time.Since(start) // time.Since = time.Now() - start
		fmt.Printf("%s %s %d %v\n", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// statusRecorder wraps http.ResponseWriter to capture the status code
// This is called EMBEDDING — ResponseWriter methods are "inherited" automatically
// We only override WriteHeader to capture the status code
type statusRecorder struct {
	http.ResponseWriter          // Embedded field — all ResponseWriter methods are available
	statusCode          int     // Our extra field to store the status code
}

// WriteHeader overrides the embedded ResponseWriter's WriteHeader
// This is called when handler does w.WriteHeader(404) etc.
func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code                  // Capture the code
	sr.ResponseWriter.WriteHeader(code)   // Pass it to the real ResponseWriter
}

// Recoverer middleware catches panics and returns 500 instead of crashing the server
// Without this, a panic in any handler would kill the entire server process!
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer + recover is Go's way of handling panics (like try/catch in other languages)
		// defer = run when function exits
		// recover() = catches a panic and returns its value, or nil if no panic
		defer func() {
			if err := recover(); err != nil {
				// Log the panic for debugging
				fmt.Printf("PANIC: %s %s — %v\n", r.Method, r.URL.Path, err)

				// Return 500 Internal Server Error to the client
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(500)
				fmt.Fprintf(w, `{"error":"internal server error"}`)
			}
		}()

		// Call the next handler — if it panics, the defer above catches it
		next.ServeHTTP(w, r)
	})
}
