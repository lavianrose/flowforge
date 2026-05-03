package integration

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lavianrose/flowforge/internal/auth"
	"github.com/lavianrose/flowforge/internal/config"
	"github.com/lavianrose/flowforge/internal/db"
	"github.com/lavianrose/flowforge/internal/migrate"
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

	// Run migrations using actual migration files
	// Resolve path: this file is at backend/tests/integration/setup.go,
	// migrations are at backend/migrations/
	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")
	if err := migrate.Up(migrationsDir); err != nil {
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

// DoRequest sends a request through the test server with a generous timeout
// to avoid flakes in CI environments where the database is slower.
func (ts *TestSuite) DoRequest(req *http.Request) (*http.Response, error) {
	return ts.Server.GetApp().Test(req, 10000) // 10s timeout for CI
}
