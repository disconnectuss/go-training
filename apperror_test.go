package main

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

// Step 8 Test: AppError implements the error interface
func TestAppErrorImplementsError(t *testing.T) {
	err := ErrNotFound("user")

	// AppError.Error() should return the message string
	if err.Error() != "user not found" {
		t.Errorf("expected 'user not found', got '%s'", err.Error())
	}

	// AppError should carry the HTTP status code
	if err.Code != 404 {
		t.Errorf("expected code 404, got %d", err.Code)
	}
}

// Step 8 Test: Different error types have different status codes
func TestAppErrorTypes(t *testing.T) {
	// Table-driven test — a common Go testing pattern
	// Instead of writing separate test functions, we define test cases in a slice
	tests := []struct {
		name    string   // Test case name (for readability in output)
		err     *AppError
		code    int
		message string
	}{
		{"not found", ErrNotFound("user"), 404, "user not found"},
		{"bad request", ErrBadRequest("invalid JSON"), 400, "invalid JSON"},
		{"internal", ErrInternal("database error"), 500, "database error"},
	}

	// range over test cases — each one runs as a sub-test
	for _, tt := range tests {
		// t.Run creates a named sub-test — appears as TestAppErrorTypes/not_found etc.
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("expected code %d, got %d", tt.code, tt.err.Code)
			}
			if tt.err.Message != tt.message {
				t.Errorf("expected message '%s', got '%s'", tt.message, tt.err.Message)
			}
		})
	}
}

// Step 8 Test: AppError can be used as a regular error (interface compatibility)
func TestAppErrorAsError(t *testing.T) {
	// This proves AppError satisfies the error interface
	var err error = ErrNotFound("user") // Assign *AppError to error variable

	// errors.As extracts a specific error type from the chain
	// It's the recommended way to check error types in Go
	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Error("expected to extract AppError from error")
	}

	if appErr.Code != 404 {
		t.Errorf("expected code 404, got %d", appErr.Code)
	}
}

// Step 8 Test: writeError sends correct status code and JSON body
func TestWriteErrorWithAppError(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, ErrNotFound("user"))

	if rec.Code != 404 {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "user not found") {
		t.Errorf("expected 'user not found', got %s", body)
	}
}

// Step 8 Test: writeError with a generic error defaults to 500
func TestWriteErrorWithGenericError(t *testing.T) {
	rec := httptest.NewRecorder()

	// A plain error (not AppError) — writeError should default to 500
	writeError(rec, errors.New("something broke"))

	if rec.Code != 500 {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "internal server error") {
		t.Errorf("expected 'internal server error', got %s", body)
	}
}
