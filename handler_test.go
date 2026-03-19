package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Step 7: MockStore — a fake UserStore for testing
// It implements the SAME interface as SQLiteStore, but uses a simple slice
// No database needed! Tests run faster and are more predictable
type MockStore struct {
	users  []User
	nextID int
}

// NewMockStore creates a MockStore with some test data
func NewMockStore() *MockStore {
	return &MockStore{
		users: []User{
			{ID: 1, Name: "Fatma", Age: 25, City: "Istanbul"},
			{ID: 2, Name: "Ahmet", Age: 30, City: "Ankara"},
		},
		nextID: 3,
	}
}

// MockStore implements ALL methods of UserStore interface
// This is what makes Go interfaces powerful — implicit implementation

func (m *MockStore) GetAll(city string) ([]User, error) {
	if city != "" {
		var filtered []User
		for _, u := range m.users {
			if u.City == city {
				filtered = append(filtered, u)
			}
		}
		if filtered == nil {
			filtered = []User{}
		}
		return filtered, nil
	}
	return m.users, nil
}

func (m *MockStore) GetByID(id int) (User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	// Step 8: Return AppError so handler gets the correct status code (404)
	return User{}, ErrNotFound("user")
}

func (m *MockStore) Create(user User) (User, error) {
	user.ID = m.nextID
	m.nextID++
	m.users = append(m.users, user)
	return user, nil
}

func (m *MockStore) Update(id int, updated User) (User, error) {
	for i := range m.users {
		if m.users[i].ID == id {
			if updated.Name != "" {
				m.users[i].Name = updated.Name
			}
			if updated.Age != 0 {
				m.users[i].Age = updated.Age
			}
			if updated.City != "" {
				m.users[i].City = updated.City
			}
			return m.users[i], nil
		}
	}
	return User{}, ErrNotFound("user")
}

func (m *MockStore) Delete(id int) (string, error) {
	for i, u := range m.users {
		if u.ID == id {
			m.users = append(m.users[:i], m.users[i+1:]...)
			return u.Name, nil
		}
	}
	return "", ErrNotFound("user")
}

// setupTestRouter creates a Chi router with a MOCK store
// No database, no disk, no network — pure in-memory testing
func setupTestRouter() *chi.Mux {
	mock := NewMockStore()
	h := NewHandler(mock) // Inject the mock!

	r := chi.NewRouter()
	r.Get("/users", h.getUsers)
	r.Post("/users", h.createUser)
	r.Get("/users/{id}", h.getUserByID)
	r.Put("/users/{id}", h.updateUser)
	r.Delete("/users/{id}", h.deleteUser)
	return r
}

// --- All existing tests now use MockStore instead of real SQLite ---

func TestGetUsers(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Fatma") {
		t.Errorf("expected 'Fatma', got %s", body)
	}
	if !strings.Contains(body, "Ahmet") {
		t.Errorf("expected 'Ahmet', got %s", body)
	}
}

func TestGetUsersFilterByCity(t *testing.T) {
	req := httptest.NewRequest("GET", "/users?city=Istanbul", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

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
	jsonBody := `{"name":"Zeynep","age":22,"city":"Izmir"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Zeynep") {
		t.Errorf("expected 'Zeynep', got %s", body)
	}
}

func TestCreateUserWithoutName(t *testing.T) {
	jsonBody := `{"age":22,"city":"Izmir"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGetUserByID(t *testing.T) {
	req := httptest.NewRequest("GET", "/users/1", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Fatma") {
		t.Errorf("expected 'Fatma', got %s", body)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/users/99", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdateUser(t *testing.T) {
	jsonBody := `{"name":"Fatma Updated","city":"Bursa"}`
	req := httptest.NewRequest("PUT", "/users/1", strings.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

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
}

func TestUpdateUserNotFound(t *testing.T) {
	jsonBody := `{"name":"Ghost"}`
	req := httptest.NewRequest("PUT", "/users/99", strings.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdateUserInvalidJSON(t *testing.T) {
	req := httptest.NewRequest("PUT", "/users/1", strings.NewReader("not json"))
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDeleteUser(t *testing.T) {
	req := httptest.NewRequest("DELETE", "/users/1", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Fatma") {
		t.Errorf("expected deleted user name 'Fatma', got %s", body)
	}
}

func TestDeleteUserNotFound(t *testing.T) {
	req := httptest.NewRequest("DELETE", "/users/99", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest("PATCH", "/users", nil)
	rec := httptest.NewRecorder()

	setupTestRouter().ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}
