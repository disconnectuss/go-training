package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Step 4: Initialize the SQLite database
	// "users.db" is the file where SQLite stores data
	// If the file doesn't exist, SQLite creates it automatically
	initDB("users.db")

	r := chi.NewRouter()

	// Step 5: Register middleware — they run in order for EVERY request
	// r.Use() adds middleware to the router's chain
	// Order matters: Recoverer first (outermost), then Logger
	// Request flow: Recoverer → Logger → Route Handler → Logger logs → Recoverer catches panics
	r.Use(Recoverer)
	r.Use(Logger)

	// Public route — no auth needed
	r.Get("/hello", handleHello)

	// Step 6: Protected routes — API key required
	// r.Group creates a sub-router that shares the same base path
	// Middleware added with r.Use inside Group only applies to THIS group
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware) // Only /users routes require API key

		r.Get("/users", getUsers)
		r.Post("/users", createUser)
		r.Get("/users/{id}", getUserByID)
		r.Put("/users/{id}", updateUser)
		r.Delete("/users/{id}", deleteUser)
	})

	fmt.Println("Server running on port 8181...")
	http.ListenAndServe(":8181", r)
}
