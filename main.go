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

	r.Get("/hello", handleHello)
	r.Get("/users", getUsers)
	r.Post("/users", createUser)
	r.Get("/users/{id}", getUserByID)
	r.Put("/users/{id}", updateUser)
	r.Delete("/users/{id}", deleteUser)

	fmt.Println("Server running on port 8181...")
	http.ListenAndServe(":8181", r)
}
