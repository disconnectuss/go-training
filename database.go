package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Step 7: SQLiteStore implements the UserStore interface using SQLite
// It holds a *sql.DB connection — this is called DEPENDENCY INJECTION
// Instead of using a global variable, we pass the dependency into the struct
type SQLiteStore struct {
	db *sql.DB // The database connection is a field, not a global!
}

// NewSQLiteStore creates a new SQLiteStore and initializes the database
// This is a "constructor function" — Go doesn't have constructors, we use functions
func NewSQLiteStore(dataSourceName string) *SQLiteStore {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		age INTEGER,
		city TEXT
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	// Return a pointer to the struct — methods below use pointer receiver
	return &SQLiteStore{db: db}
}

// GetAll returns all users, or filtered by city if city is not empty
// This method has a POINTER RECEIVER: (s *SQLiteStore)
// It means: "this method belongs to *SQLiteStore"
func (s *SQLiteStore) GetAll(city string) ([]User, error) {
	var query string
	var args []interface{}

	if city != "" {
		query = "SELECT id, name, age, city FROM users WHERE city = ?"
		args = append(args, city)
	} else {
		query = "SELECT id, name, age, city FROM users"
	}

	rows, err := s.db.Query(query, args...) // s.db instead of global db
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Age, &u.City)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if users == nil {
		users = []User{}
	}
	return users, nil
}

// GetByID returns a single user by their ID
func (s *SQLiteStore) GetByID(id int) (User, error) {
	var u User
	err := s.db.QueryRow(
		"SELECT id, name, age, city FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Name, &u.Age, &u.City)
	if err != nil {
		return User{}, fmt.Errorf("user not found")
	}
	return u, nil
}

// Create inserts a new user and returns it with the generated ID
func (s *SQLiteStore) Create(user User) (User, error) {
	result, err := s.db.Exec(
		"INSERT INTO users (name, age, city) VALUES (?, ?, ?)",
		user.Name, user.Age, user.City,
	)
	if err != nil {
		return User{}, err
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)
	return user, nil
}

// Update modifies an existing user's fields (partial update)
func (s *SQLiteStore) Update(id int, updated User) (User, error) {
	// First get the existing user
	existing, err := s.GetByID(id) // We can call our own methods!
	if err != nil {
		return User{}, err
	}

	// Apply partial updates
	if updated.Name != "" {
		existing.Name = updated.Name
	}
	if updated.Age != 0 {
		existing.Age = updated.Age
	}
	if updated.City != "" {
		existing.City = updated.City
	}

	_, err = s.db.Exec(
		"UPDATE users SET name = ?, age = ?, city = ? WHERE id = ?",
		existing.Name, existing.Age, existing.City, id,
	)
	if err != nil {
		return User{}, err
	}

	return existing, nil
}

// Delete removes a user and returns their name
func (s *SQLiteStore) Delete(id int) (string, error) {
	var name string
	err := s.db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	_, err = s.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return "", err
	}

	return name, nil
}
