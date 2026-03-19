package main

import (
	"database/sql"
	"log"

	// Step 4: Import SQLite driver
	// The underscore "_" means we import it for its SIDE EFFECTS only
	// The driver registers itself with database/sql package on import
	// We never call go-sqlite3 functions directly — we use database/sql interface
	_ "github.com/mattn/go-sqlite3"
)

// db is a global database connection pool
// *sql.DB is safe for concurrent use — it manages connections automatically
var db *sql.DB

// initDB opens the database and creates the users table if it doesn't exist
func initDB(dataSourceName string) {
	var err error

	// sql.Open doesn't actually connect — it prepares the connection pool
	// "sqlite3" is the driver name registered by go-sqlite3
	db, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatal(err) // Fatal = print error + os.Exit(1)
	}

	// Ping actually tests the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Create the users table if it doesn't exist
	// TEXT, INTEGER are SQLite data types
	// AUTOINCREMENT makes id increase automatically on each INSERT
	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		age INTEGER,
		city TEXT
	);`

	// Exec runs a query that doesn't return rows (CREATE, INSERT, UPDATE, DELETE)
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}
