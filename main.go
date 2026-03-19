package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Step 7: Create the store and inject it into the handler
	// If we wanted to switch to PostgreSQL, we'd only change THIS line
	store := NewSQLiteStore("users.db")
	h := NewHandler(store)

	r := chi.NewRouter()

	r.Use(Recoverer)
	r.Use(Logger)

	// Public route
	r.Get("/hello", h.handleHello)

	// Protected routes — now using h.methodName (method on Handler struct)
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Get("/users", h.getUsers)
		r.Post("/users", h.createUser)
		r.Get("/users/{id}", h.getUserByID)
		r.Put("/users/{id}", h.updateUser)
		r.Delete("/users/{id}", h.deleteUser)
	})

	fmt.Println("Server running on port 8181...")
	http.ListenAndServe(":8181", r)
}
