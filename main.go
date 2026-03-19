package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Step 13: Load config from environment variables (set by Kubernetes ConfigMap)
	cfg := LoadConfig()

	store := NewSQLiteStore(cfg.DBPath)
	h := NewHandler(store)

	r := chi.NewRouter()

	r.Use(Recoverer)
	r.Use(Logger)

	r.Get("/hello", h.handleHello)
	r.Get("/slow", h.handleSlow) // Step 10: Takes 5s — test graceful shutdown with this

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Get("/users", h.getUsers)
		r.Post("/users", h.createUser)
		r.Get("/users/{id}", h.getUserByID)
		r.Put("/users/{id}", h.updateUser)
		r.Delete("/users/{id}", h.deleteUser)
	})

	go func() {
		fmt.Println("Background: checking database health...")
		users, _ := store.GetAll("")
		fmt.Printf("Background: database OK — %d users loaded\n", len(users))
	}()

	// Step 10: Graceful Shutdown
	//
	// BEFORE: http.ListenAndServe(":8181", r)
	//   - Blocks forever until process is killed
	//   - Ctrl+C kills immediately — active requests are dropped!
	//
	// AFTER: We control the server lifecycle manually
	//   - Listen for OS signals (Ctrl+C = SIGINT, kill = SIGTERM)
	//   - When signal received: stop accepting NEW requests
	//   - Wait for ACTIVE requests to finish (with a timeout)
	//   - Then exit cleanly

	// Create an http.Server with explicit config
	// Previously we used http.ListenAndServe which creates one internally
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
		// Timeouts prevent slow clients from holding connections forever
		ReadTimeout:  10 * time.Second, // Max time to read the full request
		WriteTimeout: 15 * time.Second, // Max time to write the full response
		IdleTimeout:  60 * time.Second, // Max time for keep-alive connections to idle
	}

	// --- Signal Handling ---
	// signal.Notify tells Go to catch these OS signals instead of killing the process
	//
	// SIGINT  = Ctrl+C in terminal
	// SIGTERM = "kill <pid>" command, also what Kubernetes sends before killing a pod
	//
	// make(chan os.Signal, 1) — buffered channel so the signal doesn't get lost
	// if the channel is unbuffered AND we're not ready to receive, the signal is dropped!
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine so it doesn't block main
	// We need main to be free to listen for the shutdown signal
	go func() {
		fmt.Printf("Server running on port %s...\n", cfg.Port)
		// ListenAndServe blocks until the server is shut down
		// When Shutdown() is called, it returns http.ErrServerClosed
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Block here until we receive a shutdown signal
	// <-quit reads from the channel — blocks until a signal arrives
	sig := <-quit
	fmt.Printf("\nReceived signal: %v\n", sig)
	fmt.Println("Shutting down gracefully...")

	// --- Graceful Shutdown ---
	// context.WithTimeout creates a context that auto-cancels after 30 seconds
	// This gives active requests 30 seconds to finish
	// If they don't finish in time, we force shutdown anyway
	//
	// ctx    = the context to pass to Shutdown()
	// cancel = a function to release resources (must always be called)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // defer ensures cancel() is called even if something panics

	// server.Shutdown does the graceful shutdown:
	// 1. Closes the listener (no new connections accepted)
	// 2. Waits for active requests to complete
	// 3. Returns nil on success, or ctx.Err() if timeout exceeded
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Forced shutdown: %v\n", err)
	}

	fmt.Println("Server stopped cleanly.")
}
