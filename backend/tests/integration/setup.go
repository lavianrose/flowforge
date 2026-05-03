package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lavianrose/flowforge/internal/auth"
	"github.com/lavianrose/flowforge/internal/config"
	"github.com/lavianrose/flowforge/internal/db"
	"github.com/lavianrose/flowforge/internal/server"
)

type TestSuite struct {
	Server     *server.Server
	DBPool     *pgxpool.Pool
	JWTManager *auth.JWTManager
	TestTenant string
	TestUsers  map[string]string // role -> user_id
}

func Setup(t *testing.T) *TestSuite {
	// Load test config - use environment variables in CI, with localhost fallback for local dev
	cfg := &config.Config{
		PostgresURL: getEnv("POSTGRES_URL", "postgres://flowforge:flowforge@localhost:54322/flowforge?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "localhost:63797"),
		RedisPwd:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:     getEnvInt("REDIS_DB", 0),
		JWTSecret:   getEnv("JWT_SECRET", "test-secret-key"),
		Port:        getEnv("PORT", "3001"), // Different port for tests
		Env:         getEnv("ENV", "test"),
	}

	// Initialize database
	if err := db.Init(cfg); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create test database schema
	if err := runMigrations(t, cfg.PostgresURL); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test server
	srv := server.New(cfg)
	srv.Setup()

	return &TestSuite{
		Server:     srv,
		DBPool:     db.Pool,
		JWTManager: auth.NewJWTManager(cfg.JWTSecret),
		TestUsers:  make(map[string]string),
	}
}

func (ts *TestSuite) Teardown(t *testing.T) {
	// Wait for in-flight workflow executions to finish before closing DB pool
	ts.Server.WaitExecutions()

	// Clean up test data
	ctx := context.Background()
	_, err := ts.DBPool.Exec(ctx, "DELETE FROM workflow_logs")
	if err != nil {
		log.Printf("Failed to cleanup workflow_logs: %v", err)
	}

	_, err = ts.DBPool.Exec(ctx, "DELETE FROM workflow_run_steps")
	if err != nil {
		log.Printf("Failed to cleanup workflow_run_steps: %v", err)
	}

	_, err = ts.DBPool.Exec(ctx, "DELETE FROM workflow_runs")
	if err != nil {
		log.Printf("Failed to cleanup workflow_runs: %v", err)
	}

	_, err = ts.DBPool.Exec(ctx, "DELETE FROM workflow_versions")
	if err != nil {
		log.Printf("Failed to cleanup workflow_versions: %v", err)
	}

	_, err = ts.DBPool.Exec(ctx, "DELETE FROM workflows")
	if err != nil {
		log.Printf("Failed to cleanup workflows: %v", err)
	}

	_, err = ts.DBPool.Exec(ctx, "DELETE FROM schedules")
	if err != nil {
		log.Printf("Failed to cleanup schedules: %v", err)
	}

	_, err = ts.DBPool.Exec(ctx, "DELETE FROM webhooks")
	if err != nil {
		log.Printf("Failed to cleanup webhooks: %v", err)
	}

	_, err = ts.DBPool.Exec(ctx, "DELETE FROM users")
	if err != nil {
		log.Printf("Failed to cleanup users: %v", err)
	}

	_, err = ts.DBPool.Exec(ctx, "DELETE FROM tenants")
	if err != nil {
		log.Printf("Failed to cleanup tenants: %v", err)
	}

	db.Close()
}

func (ts *TestSuite) CreateTestTenant(t *testing.T) string {
	ctx := context.Background()
	var tenantID string

	err := ts.DBPool.QueryRow(ctx, `
		INSERT INTO tenants (name)
		VALUES ($1)
		RETURNING id
	`, "Test Tenant").Scan(&tenantID)

	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	ts.TestTenant = tenantID
	return tenantID
}

func (ts *TestSuite) CreateTestUser(t *testing.T, tenantID, email, password, role string) string {
	ctx := context.Background()
	var userID string

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	err = ts.DBPool.QueryRow(ctx, `
		INSERT INTO users (tenant_id, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, tenantID, email, hash, role).Scan(&userID)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	ts.TestUsers[role] = userID
	return userID
}

func (ts *TestSuite) GenerateTestToken(t *testing.T, userID, tenantID, email, role string) string {
	token, err := ts.JWTManager.Generate(userID, tenantID, email, role)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}
	return token
}

func (ts *TestSuite) CreateTestWorkflow(t *testing.T, ctx context.Context, tenantID, createdBy, name string) string {
	var workflowID string

	err := ts.DBPool.QueryRow(ctx, `
		INSERT INTO workflows (tenant_id, name, description, definition, timeout_seconds, active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, tenantID, name, "Test workflow", map[string]interface{}{
		"nodes": []map[string]interface{}{
			{
				"id":       "node-1",
				"type":     "http",
				"name":     "HTTP Request",
				"config":   map[string]string{"url": "https://api.example.com", "method": "GET"},
				"position": map[string]int{"x": 100, "y": 100},
			},
		},
		"edges": []map[string]interface{}{},
	}, 300, true, createdBy).Scan(&workflowID)

	if err != nil {
		t.Fatalf("Failed to create test workflow: %v", err)
	}

	return workflowID
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func runMigrations(t *testing.T, dbURL string) error {
	// For integration tests, we'll run the actual migration files
	// This is a simplified version - in production you'd use the migrate package
	ctx := context.Background()

	// Parse connection string
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("failed to parse db config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	// Wait for database to be ready
	for i := 0; i < 10; i++ {
		if err := pool.Ping(ctx); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Run schema creation
	_, err = pool.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'viewer',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			UNIQUE(tenant_id, email)
		);

		CREATE TABLE IF NOT EXISTS workflows (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			definition JSONB NOT NULL,
			timeout_seconds INTEGER NOT NULL DEFAULT 300,
			active BOOLEAN DEFAULT true,
			created_by UUID NOT NULL REFERENCES users(id),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS workflow_versions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			version INTEGER NOT NULL,
			definition JSONB NOT NULL,
			created_by UUID NOT NULL REFERENCES users(id),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			UNIQUE(workflow_id, version)
		);

		CREATE TABLE IF NOT EXISTS workflow_runs (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			error TEXT,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			triggered_by VARCHAR(50) DEFAULT 'manual'
		);

		CREATE TABLE IF NOT EXISTS workflow_run_steps (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
			step_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			input JSONB,
			output JSONB,
			error TEXT,
			retry_count INTEGER DEFAULT 0,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS workflow_logs (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
			step_id VARCHAR(255),
			level VARCHAR(20) NOT NULL DEFAULT 'info',
			message TEXT NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS schedules (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			cron_expression VARCHAR(100) NOT NULL,
			active BOOLEAN DEFAULT true,
			next_run_at TIMESTAMP NOT NULL,
			last_run_at TIMESTAMP,
			created_by UUID NOT NULL REFERENCES users(id),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			UNIQUE(workflow_id, tenant_id)
		);

		CREATE TABLE IF NOT EXISTS webhooks (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			path VARCHAR(255) NOT NULL UNIQUE,
			secret VARCHAR(255) NOT NULL,
			active BOOLEAN DEFAULT true,
			created_by UUID NOT NULL REFERENCES users(id),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_workflow_logs_run_id ON workflow_logs(run_id);
		CREATE INDEX IF NOT EXISTS idx_workflow_logs_created_at ON workflow_logs(created_at);
	`)

	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}
