package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Step 7: Handler struct — holds dependencies instead of using globals
// The store field is the UserStore INTERFACE, not a concrete type
// This means Handler doesn't know or care if it's SQLite, PostgreSQL, or a mock
type Handler struct {
	store UserStore // Interface field — accepts ANY type that implements UserStore
}

// NewHandler creates a Handler with the given store
// This is DEPENDENCY INJECTION: the caller decides which store to use
// Production: NewHandler(sqliteStore)
// Tests:      NewHandler(mockStore)
func NewHandler(store UserStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) handleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello! I am a Go application")
}

// Methods are now on *Handler struct — they access h.store instead of global db
func (h *Handler) getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	city := r.URL.Query().Get("city")

	// All the SQL logic is now inside the store — handler just calls the interface
	users, err := h.store.GetAll(city)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error":"database error"}`)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newUser User

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error":"invalid JSON"}`)
		return
	}

	if newUser.Name == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error":"name is required"}`)
		return
	}

	created, err := h.store.Create(newUser)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error":"database error"}`)
		return
	}

	authUser := GetUserFromContext(r)
	if authUser != "" {
		fmt.Printf("User '%s' created by API key owner: %s\n", created.Name, authUser)
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(created)
}

// parseID is still a standalone function — it doesn't need the store
func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error":"invalid id"}`)
		return 0, false
	}
	return id, true
}

func (h *Handler) getUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := parseID(w, r)
	if !ok {
		return
	}

	user, err := h.store.GetByID(id)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, `{"error":"user not found"}`)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var updated User
	err := json.NewDecoder(r.Body).Decode(&updated)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error":"invalid JSON"}`)
		return
	}

	user, err := h.store.Update(id, updated)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, `{"error":"user not found"}`)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := parseID(w, r)
	if !ok {
		return
	}

	name, err := h.store.Delete(id)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, `{"error":"user not found"}`)
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, `{"message":"user %s deleted"}`, name)
}
