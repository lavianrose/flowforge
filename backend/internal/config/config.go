package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port string
	Env  string

	// Database
	PostgresURL string

	// Redis
	RedisURL string
	RedisPwd string
	RedisDB  int

	// Auth
	JWTSecret string
}

func Load() *Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		Port:       getEnv("PORT", "3000"),
		Env:        getEnv("ENV", "development"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://localhost:5432/flowforge?sslmode=disable"),
		RedisURL:   getEnv("REDIS_URL", "localhost:6379"),
		RedisPwd:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:    redisDB,
		JWTSecret:  getEnv("JWT_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
