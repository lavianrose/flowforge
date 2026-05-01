package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type RateLimitConfig struct {
	Requests int           // Max requests
	Window   time.Duration // Time window
	KeyFunc  func(c *fiber.Ctx) string // Custom key generator
}

type RateLimiter struct {
	redisClient *redis.Client
	configs     map[string]*RateLimitConfig
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		configs:     make(map[string]*RateLimitConfig),
	}
}

func (rl *RateLimiter) AddConfig(name string, requests int, window time.Duration, keyFunc func(c *fiber.Ctx) string) {
	rl.configs[name] = &RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyFunc:  keyFunc,
	}
}

func (rl *RateLimiter) Middleware(configName string) fiber.Handler {
	config, exists := rl.configs[configName]
	if !exists {
		// Default config if not found
		config = &RateLimitConfig{
			Requests: 100,
			Window:   time.Minute,
			KeyFunc:  func(c *fiber.Ctx) string { return c.IP() },
		}
	}

	return func(c *fiber.Ctx) error {
		key := config.KeyFunc(c)
		rateLimitKey := fmt.Sprintf("ratelimit:%s", key)

		ctx := context.Background()

		// Get current count
		val, err := rl.redisClient.Get(ctx, rateLimitKey).Result()
		count := 0
		if err == nil {
			count, _ = strconv.Atoi(val)
		}

		// Check if limit exceeded
		if count >= config.Requests {
			c.Set("X-RateLimit-Limit", strconv.Itoa(config.Requests))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(config.Window).Unix(), 10))
			c.Set("Retry-After", strconv.FormatInt(int64(config.Window.Seconds()), 10))

			return c.Status(429).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		}

		// Increment counter
		pipe := rl.redisClient.Pipeline()
		incr := pipe.Incr(ctx, rateLimitKey)
		pipe.Expire(ctx, rateLimitKey, config.Window)

		_, err = pipe.Exec(ctx)
		if err != nil && err != redis.Nil {
			// If Redis fails, allow the request (fail open)
			return c.Next()
		}

		newCount := int(incr.Val())
		remaining := config.Requests - newCount

		c.Set("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(config.Window).Unix(), 10))

		return c.Next()
	}
}

// Helper functions for key generation
func ByIP(c *fiber.Ctx) string {
	return c.IP()
}

func ByUserID(c *fiber.Ctx) string {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.IP()
	}
	return fmt.Sprintf("user:%s", userID)
}

func ByTenantID(c *fiber.Ctx) string {
	tenantID := c.Locals("tenant_id")
	if tenantID == nil {
		return c.IP()
	}
	return fmt.Sprintf("tenant:%s", tenantID)
}

func ByRole(role string) func(c *fiber.Ctx) string {
	return func(c *fiber.Ctx) string {
		userRole := c.Locals("role")
		if userRole == nil {
			return c.IP()
		}
		return fmt.Sprintf("role:%s:%s", userRole, c.IP())
	}
}
