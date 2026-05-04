package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/lavianrose/flowforge/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_Auth_QueryToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	middleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	app.Use(middleware.Auth())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
			"role":    c.Locals("role"),
		})
	})

	token, err := jwtManager.Generate("user-1", "tenant-1", "test@example.com", "admin")
	require.NoError(t, err)

	// Use query parameter instead of header
	req := httptest.NewRequest("GET", "/test?token="+token, nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "user-1", body["user_id"])
	assert.Equal(t, "admin", body["role"])
}

func TestAuthMiddleware_Auth_HeaderTakesPrecedenceOverQuery(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	middleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	app.Use(middleware.Auth())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
		})
	})

	headerToken, err := jwtManager.Generate("header-user", "tenant-1", "test@example.com", "admin")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test?token=invalid-token", nil)
	req.Header.Set("Authorization", "Bearer "+headerToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "header-user", body["user_id"])
}

func TestAuthMiddleware_Auth_InvalidQueryToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	middleware := NewAuthMiddleware(jwtManager)

	app := fiber.New()
	app.Use(middleware.Auth())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test?token=invalid", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}
