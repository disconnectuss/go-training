package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Step 5 Test: Logger middleware should not change the response
// It should just log and pass through to the next handler
func TestLoggerMiddleware(t *testing.T) {
	// Create a simple handler that returns 200 with "ok"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	// Wrap it with Logger middleware
	wrapped := Logger(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	// Logger should not change the status code or body
	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %s", rec.Body.String())
	}
}

// Step 5 Test: Logger should capture the correct status code
func TestLoggerCapturesStatusCode(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	})

	wrapped := Logger(handler)

	req := httptest.NewRequest("GET", "/missing", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

// Step 5 Test: Recoverer should catch panics and return 500
func TestRecovererCatchesPanic(t *testing.T) {
	// Create a handler that PANICS
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong!")
	})

	// Wrap it with Recoverer
	wrapped := Recoverer(handler)

	req := httptest.NewRequest("GET", "/panic", nil)
	rec := httptest.NewRecorder()

	// This should NOT crash — Recoverer catches the panic
	wrapped.ServeHTTP(rec, req)

	// Should return 500 Internal Server Error
	if rec.Code != 500 {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "internal server error") {
		t.Errorf("expected 'internal server error', got %s", body)
	}
}

// Step 5 Test: Recoverer should pass through normally when no panic
func TestRecovererNoPanic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("all good"))
	})

	wrapped := Recoverer(handler)

	req := httptest.NewRequest("GET", "/ok", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "all good" {
		t.Errorf("expected 'all good', got %s", rec.Body.String())
	}
}
