package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear all relevant env vars to test defaults
	envVars := []string{
		"PORT", "ENV", "POSTGRES_URL", "REDIS_URL", "REDIS_PASSWORD",
		"REDIS_DB", "JWT_SECRET", "DOCKER_HOST", "DOCKER_PYTHON_IMAGE",
		"DOCKER_NODE_IMAGE", "DOCKER_MEMORY_MB", "DOCKER_CPU", "DOCKER_TIMEOUT_S",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	cfg := Load()

	assert.Equal(t, "3000", cfg.Port)
	assert.Equal(t, "development", cfg.Env)
	assert.Equal(t, "postgres://localhost:5432/flowforge?sslmode=disable", cfg.PostgresURL)
	assert.Equal(t, "localhost:6379", cfg.RedisURL)
	assert.Equal(t, "", cfg.RedisPwd)
	assert.Equal(t, 0, cfg.RedisDB)
	assert.Equal(t, "change-me-in-production", cfg.JWTSecret)
	assert.Equal(t, "unix:///var/run/docker.sock", cfg.Docker.Host)
	assert.Equal(t, "flowforge/runner-python:latest", cfg.Docker.PythonImage)
	assert.Equal(t, "flowforge/runner-nodejs:latest", cfg.Docker.NodeImage)
	assert.Equal(t, int64(128), cfg.Docker.DefaultMemoryMB)
	assert.Equal(t, 0.5, cfg.Docker.DefaultCPU)
	assert.Equal(t, 30, cfg.Docker.DefaultTimeoutS)
}

func TestLoad_EnvOverrides(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		getValue func(*Config) string
	}{
		{"PORT", "PORT", "8080", func(c *Config) string { return c.Port }},
		{"ENV", "ENV", "production", func(c *Config) string { return c.Env }},
		{"POSTGRES_URL", "POSTGRES_URL", "postgres://custom:5432/db", func(c *Config) string { return c.PostgresURL }},
		{"REDIS_URL", "REDIS_URL", "redis-host:6380", func(c *Config) string { return c.RedisURL }},
		{"REDIS_PASSWORD", "REDIS_PASSWORD", "secret-pass", func(c *Config) string { return c.RedisPwd }},
		{"JWT_SECRET", "JWT_SECRET", "my-jwt-secret", func(c *Config) string { return c.JWTSecret }},
		{"DOCKER_HOST", "DOCKER_HOST", "tcp://localhost:2375", func(c *Config) string { return c.Docker.Host }},
		{"DOCKER_PYTHON_IMAGE", "DOCKER_PYTHON_IMAGE", "custom/python:v2", func(c *Config) string { return c.Docker.PythonImage }},
		{"DOCKER_NODE_IMAGE", "DOCKER_NODE_IMAGE", "custom/node:v2", func(c *Config) string { return c.Docker.NodeImage }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)

			cfg := Load()
			assert.Equal(t, tt.envValue, tt.getValue(cfg))
		})
	}
}

func TestLoad_NumericEnvVars(t *testing.T) {
	t.Run("REDIS_DB", func(t *testing.T) {
		os.Setenv("REDIS_DB", "2")
		defer os.Unsetenv("REDIS_DB")
		cfg := Load()
		assert.Equal(t, 2, cfg.RedisDB)
	})

	t.Run("DOCKER_MEMORY_MB", func(t *testing.T) {
		os.Setenv("DOCKER_MEMORY_MB", "256")
		defer os.Unsetenv("DOCKER_MEMORY_MB")
		cfg := Load()
		assert.Equal(t, int64(256), cfg.Docker.DefaultMemoryMB)
	})

	t.Run("DOCKER_CPU", func(t *testing.T) {
		os.Setenv("DOCKER_CPU", "1.0")
		defer os.Unsetenv("DOCKER_CPU")
		cfg := Load()
		assert.Equal(t, 1.0, cfg.Docker.DefaultCPU)
	})

	t.Run("DOCKER_TIMEOUT_S", func(t *testing.T) {
		os.Setenv("DOCKER_TIMEOUT_S", "60")
		defer os.Unsetenv("DOCKER_TIMEOUT_S")
		cfg := Load()
		assert.Equal(t, 60, cfg.Docker.DefaultTimeoutS)
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		os.Setenv("TEST_GETENV_KEY", "myvalue")
		defer os.Unsetenv("TEST_GETENV_KEY")
		assert.Equal(t, "myvalue", getEnv("TEST_GETENV_KEY", "default"))
	})

	t.Run("returns default when not set", func(t *testing.T) {
		os.Unsetenv("TEST_GETENV_MISSING_KEY")
		assert.Equal(t, "default", getEnv("TEST_GETENV_MISSING_KEY", "default"))
	})

	t.Run("returns default when empty", func(t *testing.T) {
		os.Setenv("TEST_GETENV_EMPTY", "")
		defer os.Unsetenv("TEST_GETENV_EMPTY")
		assert.Equal(t, "default", getEnv("TEST_GETENV_EMPTY", "default"))
	})
}

func TestLoad_InvalidNumericDefaults(t *testing.T) {
	// When invalid values are provided, Atoi/ParseInt/ParseFloat return 0
	t.Run("invalid REDIS_DB", func(t *testing.T) {
		os.Setenv("REDIS_DB", "notanumber")
		defer os.Unsetenv("REDIS_DB")
		cfg := Load()
		assert.Equal(t, 0, cfg.RedisDB)
	})

	t.Run("invalid DOCKER_MEMORY_MB", func(t *testing.T) {
		os.Setenv("DOCKER_MEMORY_MB", "invalid")
		defer os.Unsetenv("DOCKER_MEMORY_MB")
		cfg := Load()
		assert.Equal(t, int64(0), cfg.Docker.DefaultMemoryMB)
	})

	t.Run("invalid DOCKER_CPU", func(t *testing.T) {
		os.Setenv("DOCKER_CPU", "invalid")
		defer os.Unsetenv("DOCKER_CPU")
		cfg := Load()
		assert.Equal(t, 0.0, cfg.Docker.DefaultCPU)
	})

	t.Run("invalid DOCKER_TIMEOUT_S", func(t *testing.T) {
		os.Setenv("DOCKER_TIMEOUT_S", "invalid")
		defer os.Unsetenv("DOCKER_TIMEOUT_S")
		cfg := Load()
		assert.Equal(t, 0, cfg.Docker.DefaultTimeoutS)
	})
}

func TestConfig_PointerNotNil(t *testing.T) {
	cfg := Load()
	require.NotNil(t, cfg)
}
