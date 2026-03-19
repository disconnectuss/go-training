package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store UserStore
}

func NewHandler(store UserStore) *Handler {
	return &Handler{store: store}
}

// Step 8: writeError is the CENTRAL error handler
// It checks if the error is an *AppError (with status code) or a generic error
// This eliminates repetitive error handling code in every handler
//
// Type assertion with "comma ok" pattern:
//   appErr, ok := err.(*AppError)
// If err is actually an *AppError → ok=true, appErr has the value
// If err is a regular error     → ok=false, appErr is nil
func writeError(w http.ResponseWriter, err error) {
	// Try to convert the error to our custom AppError type
	appErr, ok := err.(*AppError)
	if ok {
		// It's an AppError — use its status code and message
		w.WriteHeader(appErr.Code)
		fmt.Fprintf(w, `{"error":"%s"}`, appErr.Message)
	} else {
		// It's an unknown error — default to 500 Internal Server Error
		// We don't expose internal error details to the client (security!)
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error":"internal server error"}`)
	}
}

func (h *Handler) handleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello! I am a Go application")
}

// Step 10: Slow endpoint to demonstrate graceful shutdown
// When you hit this endpoint and then Ctrl+C, the server waits for it to finish
//
// r.Context() carries the request's lifecycle:
//   - Cancelled when client disconnects
//   - Cancelled when server.Shutdown() timeout expires
//
// select with ctx.Done() lets us REACT to cancellation instead of ignoring it
func (h *Handler) handleSlow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Get the request context

	select {
	case <-time.After(5 * time.Second):
		// Normal case: 5 seconds pass without cancellation
		fmt.Fprintf(w, "Slow operation completed!")

	case <-ctx.Done():
		// Context was cancelled (client disconnected or shutdown timeout)
		// ctx.Err() tells us WHY:
		//   context.Canceled         = client disconnected
		//   context.DeadlineExceeded = timeout expired
		fmt.Printf("Slow request cancelled: %v\n", ctx.Err())
		return
	}
}

// Compare BEFORE and AFTER:
//
// BEFORE (repetitive):
//   users, err := h.store.GetAll(city)
//   if err != nil {
//       w.WriteHeader(500)
//       fmt.Fprintf(w, `{"error":"database error"}`)
//       return
//   }
//
// AFTER (clean):
//   users, err := h.store.GetAll(city)
//   if err != nil {
//       writeError(w, err)  ← one line! Status code comes from AppError
//       return
//   }

func (h *Handler) getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	city := r.URL.Query().Get("city")

	users, err := h.store.GetAll(city)
	if err != nil {
		writeError(w, err) // AppError carries the right status code
		return
	}

	json.NewEncoder(w).Encode(users)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newUser User

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		writeError(w, ErrBadRequest("invalid JSON"))
		return
	}

	if newUser.Name == "" {
		writeError(w, ErrBadRequest("name is required"))
		return
	}

	created, err := h.store.Create(newUser)
	if err != nil {
		writeError(w, err)
		return
	}

	authUser := GetUserFromContext(r)
	if authUser != "" {
		fmt.Printf("User '%s' created by API key owner: %s\n", created.Name, authUser)
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(created)
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, ErrBadRequest("invalid id"))
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
		writeError(w, err) // Store returns ErrNotFound(404) or ErrInternal(500)
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
		writeError(w, ErrBadRequest("invalid JSON"))
		return
	}

	user, err := h.store.Update(id, updated)
	if err != nil {
		writeError(w, err)
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
		writeError(w, err)
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, `{"message":"user %s deleted"}`, name)
}
