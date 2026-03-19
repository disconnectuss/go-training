package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// setupAuthRouter creates a router WITH auth middleware (like production)
func setupAuthRouter() *chi.Mux {
	r := chi.NewRouter()

	// Public route
	r.Get("/hello", handleHello)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)
		r.Get("/users", getUsers)
		r.Post("/users", createUser)
		r.Get("/users/{id}", getUserByID)
		r.Put("/users/{id}", updateUser)
		r.Delete("/users/{id}", deleteUser)
	})
	return r
}

// Step 6 Test: Request WITHOUT API key should return 401
func TestAuthNoAPIKey(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("GET", "/users", nil)
	// No X-API-Key header set!
	rec := httptest.NewRecorder()

	setupAuthRouter().ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected status 401, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "missing API key") {
		t.Errorf("expected 'missing API key' error, got %s", body)
	}
}

// Step 6 Test: Request with INVALID API key should return 403
func TestAuthInvalidAPIKey(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("X-API-Key", "invalid-key-000")
	rec := httptest.NewRecorder()

	setupAuthRouter().ServeHTTP(rec, req)

	if rec.Code != 403 {
		t.Errorf("expected status 403, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "invalid API key") {
		t.Errorf("expected 'invalid API key' error, got %s", body)
	}
}

// Step 6 Test: Request with VALID API key should return 200
func TestAuthValidAPIKey(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("X-API-Key", "key-fatma-123") // Valid key!
	rec := httptest.NewRecorder()

	setupAuthRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Fatma") {
		t.Errorf("expected users list with Fatma, got %s", body)
	}
}

// Step 6 Test: Public route /hello should work WITHOUT API key
func TestPublicRouteNoAuth(t *testing.T) {
	req := httptest.NewRequest("GET", "/hello", nil)
	// No API key needed for /hello
	rec := httptest.NewRecorder()

	setupAuthRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Hello") {
		t.Errorf("expected 'Hello', got %s", body)
	}
}

// Step 6 Test: Context should carry the user name from auth
func TestGetUserFromContext(t *testing.T) {
	setupTestDB()

	// Create a handler that checks the context
	var capturedUser string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUserFromContext(r)
		w.WriteHeader(200)
	})

	// Wrap with AuthMiddleware
	r := chi.NewRouter()
	r.With(AuthMiddleware).Get("/test", testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "key-ahmet-456")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// The middleware should have set "Ahmet" in context
	if capturedUser != "Ahmet" {
		t.Errorf("expected context user 'Ahmet', got '%s'", capturedUser)
	}
}

// Step 6 Test: Create user with auth — should log the creator
func TestCreateUserWithAuth(t *testing.T) {
	setupTestDB()

	jsonBody := `{"name":"Zeynep","age":22,"city":"Izmir"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(jsonBody))
	req.Header.Set("X-API-Key", "key-admin-789")
	rec := httptest.NewRecorder()

	setupAuthRouter().ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Zeynep") {
		t.Errorf("expected 'Zeynep', got %s", body)
	}
}
