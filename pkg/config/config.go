package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBPath         string
	StoragePath    string
	MigrationsPath string
	Port           int
}

func Load() *Config {
	return &Config{
		DBPath:         getEnv("LEVO_DB_PATH", "./data/levo.db"),
		StoragePath:    getEnv("LEVO_STORAGE_PATH", "./storage"),
		MigrationsPath: getEnv("LEVO_MIGRATIONS_PATH", "./migrations"),
		Port:           getEnvAsInt("LEVO_PORT", 8080),
	}
}

func getEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err != nil {
			return intVal
		}
	}

	return defaultValue
}
