package main

import (
	"context"
	"fmt"
	"net/http"
)

// Step 6: Auth Middleware — protects routes with API Key authentication

// validAPIKeys is a map of valid API keys to their owner names
// In production, these would come from a database or environment variables!
// map[string]string = key:value pairs where both are strings
var validAPIKeys = map[string]string{
	"key-fatma-123": "Fatma",
	"key-ahmet-456": "Ahmet",
	"key-admin-789": "Admin",
}

// contextKey is a custom type for context keys
// Why a custom type? To avoid collisions with other packages using context
// If two packages both use string("user"), they would overwrite each other
// With a custom type, only THIS package can set/get this key
type contextKey string

const (
	// These are the keys we use to store values in the request context
	contextKeyAPIKey contextKey = "apiKey"
	contextKeyUser   contextKey = "user"
)

// AuthMiddleware checks for a valid API key in the request header
// Header format: X-API-Key: key-fatma-123
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the API key from the request header
		// Headers are key-value pairs sent with every HTTP request
		apiKey := r.Header.Get("X-API-Key")

		// No API key provided
		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(401) // 401 = Unauthorized
			fmt.Fprintf(w, `{"error":"missing API key — set X-API-Key header"}`)
			return // IMPORTANT: return here to stop the chain!
		}

		// Check if the API key exists in our valid keys map
		// The "comma ok" idiom: value, exists := map[key]
		// exists is true if the key was found, false otherwise
		userName, exists := validAPIKeys[apiKey]
		if !exists {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(403) // 403 = Forbidden (key exists but invalid)
			fmt.Fprintf(w, `{"error":"invalid API key"}`)
			return
		}

		// API key is valid! Store the user info in the request CONTEXT
		// Context is like a request-scoped bag — it carries values through the handler chain
		// context.WithValue creates a NEW context with the added value (contexts are immutable)
		ctx := r.Context()                                     // Get current context
		ctx = context.WithValue(ctx, contextKeyAPIKey, apiKey)  // Add API key to context
		ctx = context.WithValue(ctx, contextKeyUser, userName)  // Add user name to context

		// r.WithContext creates a new request with the updated context
		// The original request is not modified (immutability!)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext is a helper to extract the authenticated user name from context
// Other handlers can call this to know WHO made the request
func GetUserFromContext(r *http.Request) string {
	// Type assertion: context stores values as "any" (interface{})
	// We need to assert it back to string
	// The "comma ok" pattern prevents panics if the value is missing or wrong type
	userName, ok := r.Context().Value(contextKeyUser).(string)
	if !ok {
		return "" // No user in context (unauthenticated request)
	}
	return userName
}
