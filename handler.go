package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func handleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello! I am a Go application")
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		getUsers(w, r)
	case "POST":
		createUser(w, r)
	default:
		w.WriteHeader(405)
		fmt.Fprintf(w, `{"error":"method not allowed"}`)
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")

	if city != "" {
		var filtered []User
		for _, user := range users {
			if user.City == city {
				filtered = append(filtered, user)
			}
		}
		json.NewEncoder(w).Encode(filtered)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
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

	newUser.ID = nextID
	nextID++
	users = append(users, newUser)

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(newUser)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error":"invalid id"}`)
		return
	}

	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			w.WriteHeader(200)
			fmt.Fprintf(w, `{"message":"user %s deleted"}`, user.Name)
			return
		}
	}

	w.WriteHeader(404)
	fmt.Fprintf(w, `{"error":"user not found"}`)
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "DELETE":
		deleteUser(w, r)
	default:
		w.WriteHeader(405)
		fmt.Fprintf(w, `{"error":"method not allowed"}`)
	}
}
