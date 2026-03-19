package main

// User represents a person in our system
// Step 4: Data now lives in SQLite, not in memory
// The struct is still used for JSON encoding/decoding and SQL scanning
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}
