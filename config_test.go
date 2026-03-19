package main

import (
	"os"
	"testing"
)

// Step 13 Test: LoadConfig uses default values when env vars are not set
func TestLoadConfigDefaults(t *testing.T) {
	// Clear env vars to test defaults
	os.Unsetenv("APP_PORT")
	os.Unsetenv("DB_PATH")

	cfg := LoadConfig()

	if cfg.Port != "8181" {
		t.Errorf("expected default port '8181', got '%s'", cfg.Port)
	}
	if cfg.DBPath != "users.db" {
		t.Errorf("expected default db path 'users.db', got '%s'", cfg.DBPath)
	}
}

// Step 13 Test: LoadConfig reads from environment variables
func TestLoadConfigFromEnv(t *testing.T) {
	// t.Setenv sets an env var for THIS test only — automatically cleaned up after
	// This is safer than os.Setenv because it doesn't leak to other tests
	t.Setenv("APP_PORT", "9090")
	t.Setenv("DB_PATH", "/app/data/prod.db")

	cfg := LoadConfig()

	if cfg.Port != "9090" {
		t.Errorf("expected port '9090', got '%s'", cfg.Port)
	}
	if cfg.DBPath != "/app/data/prod.db" {
		t.Errorf("expected db path '/app/data/prod.db', got '%s'", cfg.DBPath)
	}
}

// Step 13 Test: getEnv returns default when env var is empty
func TestGetEnvDefault(t *testing.T) {
	os.Unsetenv("NONEXISTENT_VAR")

	result := getEnv("NONEXISTENT_VAR", "fallback")

	if result != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", result)
	}
}

// Step 13 Test: getEnv returns env var value when set
func TestGetEnvSet(t *testing.T) {
	t.Setenv("MY_VAR", "hello")

	result := getEnv("MY_VAR", "default")

	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}
