package main

// User represents a person in our system
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

// Auto-incrementing ID counter
var nextID = 5

// Global user list
var users = []User{
	{ID: 1, Name: "Fatma", Age: 25, City: "Istanbul"},
	{ID: 2, Name: "Ahmet", Age: 30, City: "Ankara"},
	{ID: 3, Name: "Elif", Age: 22, City: "Istanbul"},
	{ID: 4, Name: "Can", Age: 28, City: "Izmir"},
}
