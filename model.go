package main

// User represents a person in our system
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

// Step 7: UserStore INTERFACE — defines WHAT operations are available, not HOW they work
// Any type that has ALL these methods automatically implements UserStore
// This is Go's "implicit interface" — no "implements" keyword needed!
//
// Why interface?
// - Handler doesn't care if data is in SQLite, PostgreSQL, or memory
// - Tests can use a fake/mock store instead of a real database
// - Easy to swap database without changing handler code
type UserStore interface {
	GetAll(city string) ([]User, error)       // List users, optionally filter by city
	GetByID(id int) (User, error)             // Get a single user
	Create(user User) (User, error)           // Create and return with generated ID
	Update(id int, user User) (User, error)   // Update and return the updated user
	Delete(id int) (string, error)            // Delete and return the deleted user's name
}
