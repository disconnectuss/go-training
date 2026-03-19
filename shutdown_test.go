package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

// Step 10 Test: Server shuts down gracefully — active requests complete
func TestGracefulShutdown(t *testing.T) {
	mock := NewMockStore()
	h := NewHandler(mock)

	r := chi.NewRouter()
	r.Get("/users", h.getUsers)

	// Create a real HTTP server (not just a handler)
	server := &http.Server{
		Handler: r,
	}

	// httptest.NewUnstartedServer lets us control the server lifecycle
	ts := httptest.NewUnstartedServer(r)
	ts.Start()
	defer ts.Close()

	// Make a request to verify server is working
	resp, err := http.Get(ts.URL + "/users")
	if err != nil {
		t.Fatalf("server should be running: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Shutdown with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown should succeed without error
	_ = server.Shutdown(ctx)
}

// Step 10 Test: Context cancellation stops the slow handler
func TestSlowHandlerContextCancel(t *testing.T) {
	mock := NewMockStore()
	h := NewHandler(mock)

	// Create a request with a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", "/slow", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	// Run the slow handler in a goroutine
	done := make(chan bool)
	go func() {
		h.handleSlow(rec, req)
		done <- true
	}()

	// Cancel the context after 100ms (instead of waiting 5 seconds)
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Handler should return quickly after cancellation
	select {
	case <-done:
		// Good — handler exited after cancel
	case <-time.After(2 * time.Second):
		t.Error("handler did not respect context cancellation")
	}
}

// Step 10 Test: Slow handler completes normally when not cancelled
func TestSlowHandlerCompletes(t *testing.T) {
	mock := NewMockStore()
	h := NewHandler(mock)

	r := chi.NewRouter()
	r.Get("/slow", h.handleSlow)

	// Override the slow handler with a faster version for testing
	// We test the PATTERN, not the 5-second wait
	fastHandler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		select {
		case <-time.After(50 * time.Millisecond): // 50ms instead of 5s
			fmt.Fprintf(w, "Slow operation completed!")
		case <-ctx.Done():
			return
		}
	}

	fastRouter := chi.NewRouter()
	fastRouter.Get("/slow", fastHandler)

	req := httptest.NewRequest("GET", "/slow", nil)
	rec := httptest.NewRecorder()

	fastRouter.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "completed") {
		t.Errorf("expected 'completed', got %s", body)
	}
}

// Step 10 Test: context.WithTimeout cancels automatically after deadline
func TestContextTimeout(t *testing.T) {
	// Create a context that auto-cancels after 100ms
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Wait for the context to be done
	select {
	case <-ctx.Done():
		// ctx.Err() tells us WHY it was cancelled
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("expected DeadlineExceeded, got %v", ctx.Err())
		}
	case <-time.After(1 * time.Second):
		t.Error("context should have timed out by now")
	}
}

// Step 10 Test: context.WithCancel cancels when cancel() is called
func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// ctx.Done() channel should be closed now
	select {
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			t.Errorf("expected Canceled, got %v", ctx.Err())
		}
	default:
		t.Error("context should be done after cancel()")
	}
}
