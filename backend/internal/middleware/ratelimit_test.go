package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestByIP(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		key := ByIP(c)
		return c.SendString(key)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestByUserID_WithUserID(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		key := ByUserID(c)
		return c.SendString(key)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	body := make([]byte, 100)
	n, _ := resp.Body.Read(body)
	assert.Equal(t, "user:user-123", string(body[:n]))
}

func TestByUserID_FallbackToIP(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		key := ByUserID(c)
		return c.SendString(key)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestByTenantID_WithTenantID(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("tenant_id", "tenant-456")
		key := ByTenantID(c)
		return c.SendString(key)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	body := make([]byte, 100)
	n, _ := resp.Body.Read(body)
	assert.Equal(t, "tenant:tenant-456", string(body[:n]))
}

func TestByTenantID_FallbackToIP(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		key := ByTenantID(c)
		return c.SendString(key)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestByRole_WithRole(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("role", "admin")
		keyFn := ByRole("admin")
		return c.SendString(keyFn(c))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	body := make([]byte, 200)
	n, _ := resp.Body.Read(body)
	result := string(body[:n])
	assert.Contains(t, result, "role:admin:")
}

func TestByRole_FallbackToIP(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		keyFn := ByRole("admin")
		return c.SendString(keyFn(c))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestRateLimiter_AddConfig(t *testing.T) {
	rl := NewRateLimiter(nil)
	rl.AddConfig("test", 100, 0, ByIP)

	assert.NotNil(t, rl.configs["test"])
	assert.Equal(t, 100, rl.configs["test"].Requests)
	assert.NotNil(t, rl.configs["test"].KeyFunc)
}

func TestRateLimiter_AddConfig_Multiple(t *testing.T) {
	rl := NewRateLimiter(nil)
	rl.AddConfig("auth", 10, 0, ByIP)
	rl.AddConfig("api", 100, 0, ByUserID)

	assert.Equal(t, 10, rl.configs["auth"].Requests)
	assert.Equal(t, 100, rl.configs["api"].Requests)
}

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(nil)
	assert.NotNil(t, rl)
	assert.NotNil(t, rl.configs)
	assert.Empty(t, rl.configs)
}
