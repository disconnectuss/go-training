package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func handleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello! I am a Go application")
}

// getUsers returns all users, or filtered by city query parameter
// Step 4: Now reads from SQLite instead of the in-memory slice
func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	city := r.URL.Query().Get("city")

	// Build the SQL query based on whether city filter is provided
	var query string
	var args []interface{} // args holds the query parameters (prevents SQL injection!)

	if city != "" {
		// "?" is a placeholder — SQLite replaces it with the value safely
		query = "SELECT id, name, age, city FROM users WHERE city = ?"
		args = append(args, city)
	} else {
		query = "SELECT id, name, age, city FROM users"
	}

	// db.Query runs a SELECT and returns multiple rows
	rows, err := db.Query(query, args...)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error":"database error"}`)
		return
	}
	// defer = run this when the function exits (cleanup)
	// Always close rows to free the database connection!
	defer rows.Close()

	// Scan each row into a User struct
	var userList []User
	for rows.Next() { // Next() moves to the next row, returns false when done
		var u User
		// Scan reads column values INTO the variables (order must match SELECT)
		err := rows.Scan(&u.ID, &u.Name, &u.Age, &u.City)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, `{"error":"scan error"}`)
			return
		}
		userList = append(userList, u)
	}

	// Return empty array instead of null when no users found
	if userList == nil {
		userList = []User{}
	}

	json.NewEncoder(w).Encode(userList)
}

// createUser adds a new user to the database
// Step 6: Now logs WHO created the user using context
func createUser(w http.ResponseWriter, r *http.Request) {
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

	// db.Exec for INSERT — returns a Result with LastInsertId and RowsAffected
	result, err := db.Exec(
		"INSERT INTO users (name, age, city) VALUES (?, ?, ?)",
		newUser.Name, newUser.Age, newUser.City,
	)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error":"database error"}`)
		return
	}

	// Step 6: Log who created the user (from auth context)
	authUser := GetUserFromContext(r)
	if authUser != "" {
		fmt.Printf("User '%s' created by API key owner: %s\n", newUser.Name, authUser)
	}

	// Get the auto-generated ID from SQLite
	id, _ := result.LastInsertId()
	newUser.ID = int(id) // LastInsertId returns int64, we convert to int

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(newUser)
}

// parseID extracts and converts the {id} URL parameter
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

// getUserByID returns a single user from the database
func getUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var u User
	// QueryRow returns a single row — use when you expect 0 or 1 result
	err := db.QueryRow(
		"SELECT id, name, age, city FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Name, &u.Age, &u.City)

	if err != nil {
		// sql.ErrNoRows means the query returned 0 rows
		w.WriteHeader(404)
		fmt.Fprintf(w, `{"error":"user not found"}`)
		return
	}

	json.NewEncoder(w).Encode(u)
}

// updateUser updates an existing user in the database
func updateUser(w http.ResponseWriter, r *http.Request) {
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

	// First check if the user exists
	var existing User
	err = db.QueryRow(
		"SELECT id, name, age, city FROM users WHERE id = ?", id,
	).Scan(&existing.ID, &existing.Name, &existing.Age, &existing.City)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, `{"error":"user not found"}`)
		return
	}

	// Apply partial updates — only change fields that were sent
	if updated.Name != "" {
		existing.Name = updated.Name
	}
	if updated.Age != 0 {
		existing.Age = updated.Age
	}
	if updated.City != "" {
		existing.City = updated.City
	}

	// UPDATE the row in the database
	_, err = db.Exec(
		"UPDATE users SET name = ?, age = ?, city = ? WHERE id = ?",
		existing.Name, existing.Age, existing.City, id,
	)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error":"database error"}`)
		return
	}

	json.NewEncoder(w).Encode(existing)
}

// deleteUser removes a user from the database
func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := parseID(w, r)
	if !ok {
		return
	}

	// First get the user name for the response message
	var name string
	err := db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, `{"error":"user not found"}`)
		return
	}

	// DELETE the row
	_, err = db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error":"database error"}`)
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, `{"message":"user %s deleted"}`, name)
}
