package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// setupAuthRouter creates a router WITH auth middleware using MockStore
func setupAuthRouter() *chi.Mux {
	mock := NewMockStore()
	h := NewHandler(mock)

	r := chi.NewRouter()
	r.Get("/hello", h.handleHello)
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)
		r.Get("/users", h.getUsers)
		r.Post("/users", h.createUser)
		r.Get("/users/{id}", h.getUserByID)
		r.Put("/users/{id}", h.updateUser)
		r.Delete("/users/{id}", h.deleteUser)
	})
	return r
}

func TestAuthNoAPIKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
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

func TestAuthInvalidAPIKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("X-API-Key", "invalid-key-000")
	rec := httptest.NewRecorder()

	setupAuthRouter().ServeHTTP(rec, req)

	if rec.Code != 403 {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestAuthValidAPIKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("X-API-Key", "key-fatma-123")
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

func TestPublicRouteNoAuth(t *testing.T) {
	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()

	setupAuthRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetUserFromContext(t *testing.T) {
	var capturedUser string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUserFromContext(r)
		w.WriteHeader(200)
	})

	r := chi.NewRouter()
	r.With(AuthMiddleware).Get("/test", testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "key-ahmet-456")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if capturedUser != "Ahmet" {
		t.Errorf("expected context user 'Ahmet', got '%s'", capturedUser)
	}
}

func TestCreateUserWithAuth(t *testing.T) {
	jsonBody := `{"name":"Zeynep","age":22,"city":"Izmir"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(jsonBody))
	req.Header.Set("X-API-Key", "key-admin-789")
	rec := httptest.NewRecorder()

	setupAuthRouter().ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}
