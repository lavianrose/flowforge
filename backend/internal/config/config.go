package config

import (
	"os"
	"strconv"
)

type DockerConfig struct {
	Host            string
	PythonImage     string
	NodeImage       string
	DefaultMemoryMB int64
	DefaultCPU      float64
	DefaultTimeoutS int
}

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

	// Docker
	Docker DockerConfig
}

func Load() *Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	memoryMB, _ := strconv.ParseInt(getEnv("DOCKER_MEMORY_MB", "128"), 10, 64)
	cpu, _ := strconv.ParseFloat(getEnv("DOCKER_CPU", "0.5"), 64)
	timeoutS, _ := strconv.Atoi(getEnv("DOCKER_TIMEOUT_S", "30"))

	return &Config{
		Port:        getEnv("PORT", "3000"),
		Env:         getEnv("ENV", "development"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://localhost:5432/flowforge?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		RedisPwd:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:     redisDB,
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
		Docker: DockerConfig{
			Host:            getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
			PythonImage:     getEnv("DOCKER_PYTHON_IMAGE", "flowforge/runner-python:latest"),
			NodeImage:       getEnv("DOCKER_NODE_IMAGE", "flowforge/runner-nodejs:latest"),
			DefaultMemoryMB: memoryMB,
			DefaultCPU:      cpu,
			DefaultTimeoutS: timeoutS,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
