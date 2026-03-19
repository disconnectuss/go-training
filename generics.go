package main

import "fmt"

// Step 12: Generics — write ONE function that works with MANY types
//
// BEFORE generics (Go < 1.18): You had to write separate functions:
//   func ContainsInt(slice []int, val int) bool { ... }
//   func ContainsString(slice []string, val string) bool { ... }
//   func ContainsFloat(slice []float64, val float64) bool { ... }
//   → Same logic, copy-pasted 3 times!
//
// WITH generics: Write it ONCE:
//   func Contains[T comparable](slice []T, val T) bool { ... }
//   → Works with int, string, float64, and any comparable type!

// --- Example 1: Generic function ---

// Contains checks if a slice contains a value
// [T comparable] is a TYPE PARAMETER:
//   T          = placeholder for any type (like a variable for types)
//   comparable = CONSTRAINT — T must support == and != operators
//                (int, string, float64, bool, structs are comparable)
//                (slices, maps, functions are NOT comparable)
func Contains[T comparable](slice []T, val T) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// --- Example 2: Generic function with multiple type parameters ---

// Map transforms a slice by applying a function to each element
// [T any, R any] = two type parameters, both can be ANY type
//   T = input type
//   R = result type (can be different from T!)
//
// "any" is an alias for "interface{}" — it means NO constraint (accepts everything)
func Map[T any, R any](slice []T, fn func(T) R) []R {
	result := make([]R, len(slice)) // Allocate result slice with same length
	for i, item := range slice {
		result[i] = fn(item) // Apply the function to each element
	}
	return result
}

// --- Example 3: Generic function — Filter ---

// Filter returns only elements that satisfy the predicate function
// predicate = a function that returns true/false
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// --- Example 4: Type constraint with interface ---

// Number is a CUSTOM CONSTRAINT — a union of numeric types
// The "|" means OR: T can be int OR float64 OR int64 etc.
// "~int" means "any type whose UNDERLYING type is int"
//   (this includes custom types like "type Age int")
type Number interface {
	~int | ~int64 | ~float64
}

// Sum adds all numbers in a slice
// Works with int, int64, float64, and any type based on them
func Sum[T Number](numbers []T) T {
	var total T // Zero value of T (0 for numbers)
	for _, n := range numbers {
		total += n
	}
	return total
}

// --- Example 5: Generic struct ---

// Pair holds two values of potentially DIFFERENT types
// Like a tuple in Python: (1, "hello")
type Pair[T any, U any] struct {
	First  T
	Second U
}

// NewPair is a constructor — Go can INFER type parameters from arguments
// NewPair(1, "hello") → Go infers T=int, U=string automatically
func NewPair[T any, U any](first T, second U) Pair[T, U] {
	return Pair[T, U]{First: first, Second: second}
}

// String method on generic struct — uses pointer receiver
func (p Pair[T, U]) String() string {
	return fmt.Sprintf("(%v, %v)", p.First, p.Second)
}

// --- Example 6: Generic Response wrapper (practical for our API!) ---

// Response wraps any data type with a success/error status
// This is useful for API responses — same structure, different data types
type Response[T any] struct {
	Data    T      `json:"data"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"` // omitempty = skip if empty string
}

// NewSuccessResponse creates a success response with data
func NewSuccessResponse[T any](data T) Response[T] {
	return Response[T]{
		Data:    data,
		Success: true,
	}
}

// NewErrorResponse creates an error response
// We use "any" for data because error responses usually have no data
func NewErrorResponse(message string) Response[any] {
	return Response[any]{
		Success: false,
		Message: message,
	}
}

// --- Example 7: Keys — extract keys from any map ---

// Keys returns all keys from a map
// [K comparable, V any] — K must be comparable (map keys must be)
//                          V can be anything
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m)) // Pre-allocate with capacity
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
