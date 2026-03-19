package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// setupTestDB creates a fresh in-memory SQLite database for each test
// ":memory:" means the database lives in RAM — it's destroyed when the test ends
// This ensures each test starts with a clean state
func setupTestDB() {
	initDB(":memory:")

	// Insert test data
	db.Exec("INSERT INTO users (name, age, city) VALUES (?, ?, ?)", "Fatma", 25, "Istanbul")
	db.Exec("INSERT INTO users (name, age, city) VALUES (?, ?, ?)", "Ahmet", 30, "Ankara")
}

// setupRouter creates a Chi router with all routes for testing
func setupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/users", getUsers)
	r.Post("/users", createUser)
	r.Get("/users/{id}", getUserByID)
	r.Put("/users/{id}", updateUser)
	r.Delete("/users/{id}", deleteUser)
	return r
}

func TestGetUsers(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Fatma") {
		t.Errorf("expected body to contain 'Fatma', got %s", body)
	}
	if !strings.Contains(body, "Ahmet") {
		t.Errorf("expected body to contain 'Ahmet', got %s", body)
	}
}

func TestGetUsersFilterByCity(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("GET", "/users?city=Istanbul", nil)
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Fatma") {
		t.Errorf("expected Fatma (Istanbul), got %s", body)
	}
	if strings.Contains(body, "Ahmet") {
		t.Errorf("Ahmet (Ankara) should not be in Istanbul filter, got %s", body)
	}
}

func TestCreateUser(t *testing.T) {
	setupTestDB()

	jsonBody := `{"name":"Zeynep","age":22,"city":"Izmir"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Zeynep") {
		t.Errorf("expected body to contain 'Zeynep', got %s", body)
	}

	// Verify user count in database
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count != 3 {
		t.Errorf("expected 3 users in db, got %d", count)
	}
}

func TestCreateUserWithoutName(t *testing.T) {
	setupTestDB()

	jsonBody := `{"age":22,"city":"Izmir"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	// Database should still have only 2 users
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count != 2 {
		t.Errorf("expected 2 users (unchanged), got %d", count)
	}
}

func TestGetUserByID(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("GET", "/users/1", nil)
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Fatma") {
		t.Errorf("expected body to contain 'Fatma', got %s", body)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("GET", "/users/99", nil)
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdateUser(t *testing.T) {
	setupTestDB()

	jsonBody := `{"name":"Fatma Updated","city":"Bursa"}`
	req := httptest.NewRequest("PUT", "/users/1", strings.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Fatma Updated") {
		t.Errorf("expected 'Fatma Updated', got %s", body)
	}
	if !strings.Contains(body, "Bursa") {
		t.Errorf("expected city 'Bursa', got %s", body)
	}

	// Verify the change is persisted in the database
	var name string
	db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name)
	if name != "Fatma Updated" {
		t.Errorf("expected db name 'Fatma Updated', got %s", name)
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	setupTestDB()

	jsonBody := `{"name":"Ghost"}`
	req := httptest.NewRequest("PUT", "/users/99", strings.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdateUserInvalidJSON(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("PUT", "/users/1", strings.NewReader("not json"))
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDeleteUser(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("DELETE", "/users/1", nil)
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Verify only 1 user remains
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 user after delete, got %d", count)
	}
}

func TestDeleteUserNotFound(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("DELETE", "/users/99", nil)
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	// Database should still have 2 users
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count != 2 {
		t.Errorf("expected 2 users (unchanged), got %d", count)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	setupTestDB()

	req := httptest.NewRequest("PATCH", "/users", nil)
	rec := httptest.NewRecorder()

	setupRouter().ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}
