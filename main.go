package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", handleHello)
	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/users/", handleUserByID)

	fmt.Println("Server running on port 8181...")
	http.ListenAndServe(":8181", nil)
}
