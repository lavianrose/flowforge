package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lavianrose/flowforge/internal/repository"
)

type StatsHandler struct {
	runRepo *repository.RunRepository
}

func NewStatsHandler(runRepo *repository.RunRepository) *StatsHandler {
	return &StatsHandler{
		runRepo: runRepo,
	}
}

func (h *StatsHandler) GetHealthStats(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	stats, err := h.runRepo.GetHealthStats(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch health stats"})
	}

	return c.JSON(stats)
}
