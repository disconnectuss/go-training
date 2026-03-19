package main

// User represents a person in our system
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

// UserStore defines the data operations for users
// Methods return error — which can be a regular error OR an *AppError
type UserStore interface {
	GetAll(city string) ([]User, error)
	GetByID(id int) (User, error)
	Create(user User) (User, error)
	Update(id int, user User) (User, error)
	Delete(id int) (string, error)
}
