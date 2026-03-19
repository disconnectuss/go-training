package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

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

	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) GetAll(city string) ([]User, error) {
	var query string
	var args []interface{}

	if city != "" {
		query = "SELECT id, name, age, city FROM users WHERE city = ?"
		args = append(args, city)
	} else {
		query = "SELECT id, name, age, city FROM users"
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		// Step 8: Return AppError with proper status code instead of raw error
		return nil, ErrInternal("database error")
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Age, &u.City)
		if err != nil {
			return nil, ErrInternal("scan error")
		}
		users = append(users, u)
	}

	if users == nil {
		users = []User{}
	}
	return users, nil
}

func (s *SQLiteStore) GetByID(id int) (User, error) {
	var u User
	err := s.db.QueryRow(
		"SELECT id, name, age, city FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Name, &u.Age, &u.City)
	if err != nil {
		if err == sql.ErrNoRows {
			// Step 8: Not found → 404, not a generic error
			return User{}, ErrNotFound("user")
		}
		return User{}, ErrInternal("database error")
	}
	return u, nil
}

func (s *SQLiteStore) Create(user User) (User, error) {
	result, err := s.db.Exec(
		"INSERT INTO users (name, age, city) VALUES (?, ?, ?)",
		user.Name, user.Age, user.City,
	)
	if err != nil {
		return User{}, ErrInternal("database error")
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)
	return user, nil
}

func (s *SQLiteStore) Update(id int, updated User) (User, error) {
	existing, err := s.GetByID(id)
	if err != nil {
		return User{}, err // Already an AppError from GetByID
	}

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
		return User{}, ErrInternal("database error")
	}

	return existing, nil
}

func (s *SQLiteStore) Delete(id int) (string, error) {
	var name string
	err := s.db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound("user")
		}
		return "", ErrInternal("database error")
	}

	_, err = s.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return "", ErrInternal("database error")
	}

	return name, nil
}
