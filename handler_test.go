package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// Her testten önce user listesini sıfırla
// Böylece testler birbirini etkilemez
func resetUsers() {
	users = []User{
		{ID: 1, Name: "Fatma", Age: 25, City: "Istanbul"},
		{ID: 2, Name: "Ahmet", Age: 30, City: "Ankara"},
	}
	nextID = 3
}

func TestGetUsers(t *testing.T) {
	resetUsers()

	// 1. Fake bir HTTP request oluştur
	req := httptest.NewRequest("GET", "/users", nil)

	// 2. Fake bir ResponseWriter oluştur (cevabı yakalar)
	rec := httptest.NewRecorder()

	// 3. Handler'ı çağır
	handleUsers(rec, req)

	// 4. Kontrol et
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
	resetUsers()

	req := httptest.NewRequest("GET", "/users?city=Istanbul", nil)
	rec := httptest.NewRecorder()

	handleUsers(rec, req)

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
	resetUsers()

	jsonBody := `{"name":"Zeynep","age":22,"city":"Izmir"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	handleUsers(rec, req)

	if rec.Code != 201 {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Zeynep") {
		t.Errorf("expected body to contain 'Zeynep', got %s", body)
	}

	// User listesinde 3 kişi olmalı
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}

func TestCreateUserWithoutName(t *testing.T) {
	resetUsers()

	jsonBody := `{"age":22,"city":"Izmir"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(jsonBody))
	rec := httptest.NewRecorder()

	handleUsers(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	// Liste değişmemeli
	if len(users) != 2 {
		t.Errorf("expected 2 users (unchanged), got %d", len(users))
	}
}

func TestDeleteUser(t *testing.T) {
	resetUsers()

	req := httptest.NewRequest("DELETE", "/users/1", nil)
	rec := httptest.NewRecorder()

	handleUserByID(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if len(users) != 1 {
		t.Errorf("expected 1 user after delete, got %d", len(users))
	}

	// Kalan user Ahmet olmalı
	if users[0].Name != "Ahmet" {
		t.Errorf("expected remaining user to be Ahmet, got %s", users[0].Name)
	}
}

func TestDeleteUserNotFound(t *testing.T) {
	resetUsers()

	req := httptest.NewRequest("DELETE", "/users/99", nil)
	rec := httptest.NewRecorder()

	handleUserByID(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	// Liste değişmemeli
	if len(users) != 2 {
		t.Errorf("expected 2 users (unchanged), got %d", len(users))
	}
}

func TestMethodNotAllowed(t *testing.T) {
	resetUsers()

	req := httptest.NewRequest("PUT", "/users", nil)
	rec := httptest.NewRecorder()

	handleUsers(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}
