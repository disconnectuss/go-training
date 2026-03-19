package main

import "fmt"

// Step 8: AppError — a custom error type that carries HTTP status code
// Go's built-in error interface has only one method: Error() string
// We extend it with a status code so the handler knows what HTTP status to return
//
// The error interface:
//   type error interface {
//       Error() string
//   }
// ANY type with an Error() string method is an error in Go!

type AppError struct {
	Code    int    // HTTP status code (400, 404, 500, etc.)
	Message string // Error message to send to the client
}

// Error() makes AppError implement the built-in error interface
// Now AppError can be used anywhere a regular error is expected
func (e *AppError) Error() string {
	return e.Message
}

// Helper functions to create common errors
// These make the code more readable: return ErrNotFound("user") instead of &AppError{404, "..."}

func ErrNotFound(resource string) *AppError {
	return &AppError{
		Code:    404,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

func ErrBadRequest(message string) *AppError {
	return &AppError{
		Code:    400,
		Message: message,
	}
}

func ErrInternal(message string) *AppError {
	return &AppError{
		Code:    500,
		Message: message,
	}
}
