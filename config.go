package main

import "os"

// Step 13: Config reads settings from environment variables
// In Kubernetes, these come from ConfigMap (deployment/configmap.yaml)
// Locally, they use default values
//
// This is the 12-Factor App principle: "Store config in the environment"
// Same binary runs in dev, staging, and prod — only env vars change

// Config holds application configuration
type Config struct {
	Port   string // Server port (APP_PORT)
	DBPath string // SQLite database path (DB_PATH)
}

// LoadConfig reads config from environment variables with fallback defaults
// os.Getenv returns "" if the variable is not set
func LoadConfig() Config {
	return Config{
		Port:   getEnv("APP_PORT", "8181"),
		DBPath: getEnv("DB_PATH", "users.db"),
	}
}

// getEnv reads an environment variable or returns a default value
// This pattern avoids repeating the if/else everywhere
func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}
