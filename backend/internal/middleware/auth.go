package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/lavianrose/flowforge/internal/auth"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := ""

		// Try Authorization header first
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		// Fallback to query parameter (needed for SSE/EventSource which can't set headers)
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Missing authorization"})
		}

		claims, err := m.jwtManager.Validate(token)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func (m *AuthMiddleware) Role(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")
		if userRole == nil {
			return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
		}

		roleStr := userRole.(string)
		for _, role := range roles {
			if role == roleStr {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{"error": "Insufficient permissions"})
	}
}
