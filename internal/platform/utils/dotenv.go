package util

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// GodotEnv loads environment variables from .env file in non-production environments
// and returns the value for the given key
func GodotEnv(key string) string {
	env := make(chan string, 1)

	if os.Getenv("GO_ENV") != "production" {
		godotenv.Load(".env")
		env <- os.Getenv(key)
	} else {
		env <- os.Getenv(key)
	}

	return <-env
}

// GetEnv returns the environment variable or default value if not set
func GetEnv(key string, defaultVal string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return value
}

// GetEnvAsInt parses an environment variable as an integer
func GetEnvAsInt(key string, defaultVal int) int {
	valueStr := GodotEnv(key)
	if valueStr == "" {
		return defaultVal
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}

	return value
}

// GetEnvAsInt64 parses an environment variable as a 64-bit integer
func GetEnvAsInt64(key string, defaultVal int64) int64 {
	valueStr := GodotEnv(key)
	if valueStr == "" {
		return defaultVal
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultVal
	}

	return value
}

// GetEnvAsBool parses an environment variable as a boolean
// Values "true", "1", "yes" evaluate to true
// Values "false", "0", "no" evaluate to false
// If parsing fails, returns the default value
func GetEnvAsBool(key string, defaultVal bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}

	if valueStr == "true" || valueStr == "1" || valueStr == "yes" {
		return true
	}
	if valueStr == "false" || valueStr == "0" || valueStr == "no" {
		return false
	}

	return defaultVal
}

// GetEnvAsDuration parses an environment variable as a duration in seconds
// If parsing fails, returns the default value
func GetEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}

	intVal, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}

	return time.Duration(intVal) * time.Second
}

// GetEnvAsDurationMs parses an environment variable as a duration in milliseconds
// If parsing fails, returns the default value
func GetEnvAsDurationMs(key string, defaultVal time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}

	intVal, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}

	return time.Duration(intVal) * time.Millisecond
}
