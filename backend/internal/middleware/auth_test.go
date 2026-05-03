package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/lavianrose/flowforge/internal/auth"
)

func TestAuthMiddleware_Auth_Success(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	middleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	app.Use(middleware.Auth())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id":   c.Locals("user_id"),
			"tenant_id": c.Locals("tenant_id"),
			"email":     c.Locals("email"),
			"role":      c.Locals("role"),
		})
	})

	// Generate a valid token
	token, err := jwtManager.Generate("user-1", "tenant-1", "test@example.com", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	// Check status
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_Auth_MissingHeader(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	middleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	app.Use(middleware.Auth())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("test")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != 401 {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_Auth_InvalidHeaderFormat(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	middleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	app.Use(middleware.Auth())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("test")
	})

	tests := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "invalid-token"},
		{"Missing token", "Bearer"},
		{"Wrong prefix", "Basic token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.header)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != 401 {
				t.Errorf("Expected status 401, got %d", resp.StatusCode)
			}
		})
	}
}

func TestAuthMiddleware_Auth_InvalidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	middleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	app.Use(middleware.Auth())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("test")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != 401 {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_Role_Success(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	authMiddleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	app.Use(authMiddleware.Auth())
	app.Use(authMiddleware.Role("admin", "editor"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("authorized")
	})

	// Test with admin role
	token, err := jwtManager.Generate("user-1", "tenant-1", "test@example.com", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_Role_InsufficientPermissions(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	authMiddleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	app.Use(authMiddleware.Auth())
	app.Use(authMiddleware.Role("admin"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("authorized")
	})

	// Test with viewer role
	token, err := jwtManager.Generate("user-1", "tenant-1", "test@example.com", "viewer")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != 403 {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_Role_NotAuthenticated(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	authMiddleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	// Skip auth middleware to test role middleware without auth
	app.Use(authMiddleware.Role("admin"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("authorized")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != 401 {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_Role_MultipleRoles(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	authMiddleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	app.Use(authMiddleware.Auth())
	app.Use(authMiddleware.Role("admin", "editor", "viewer"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("authorized")
	})

	roles := []string{"admin", "editor", "viewer"}

	for _, role := range roles {
		t.Run("Role_"+role, func(t *testing.T) {
			token, err := jwtManager.Generate("user-1", "tenant-1", "test@example.com", role)
			if err != nil {
				t.Fatalf("Failed to generate token: %v", err)
			}

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != 200 {
				t.Errorf("Expected status 200 for role %s, got %d", role, resp.StatusCode)
			}
		})
	}
}
