package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lavianrose/flowforge/internal/models"
	"github.com/lavianrose/flowforge/internal/repository"
)

type RunHandler struct {
	runRepo *repository.RunRepository
}

func NewRunHandler(runRepo *repository.RunRepository) *RunHandler {
	return &RunHandler{runRepo: runRepo}
}

func (h *RunHandler) GetRun(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	id := c.Params("id")

	run, err := h.runRepo.Get(c.Context(), id, tenantID)
	if err != nil {
		if err.Error() == "run not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Run not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get run"})
	}

	steps, err := h.runRepo.GetSteps(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get steps"})
	}

	return c.JSON(fiber.Map{
		"run":   run,
		"steps": steps,
	})
}

func (h *RunHandler) StreamRun(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	id := c.Params("id")

	// Verify run exists and belongs to tenant
	run, err := h.runRepo.Get(c.Context(), id, tenantID)
	if err != nil {
		if err.Error() == "run not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Run not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get run"})
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")

	// Send initial state
	h.sendEvent(c, "run_state", run)

	// Get initial steps
	initialSteps, err := h.runRepo.GetSteps(c.Context(), id)
	if err != nil {
		h.sendError(c, "Failed to get initial steps")
		return nil
	}
	h.sendEvent(c, "steps_state", initialSteps)

	// Poll for updates (in production, use Redis pub/sub or WebSocket)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Context().Done():
			return nil
		case <-ticker.C:
			// Check run status
			currentRun, err := h.runRepo.Get(c.Context(), id, tenantID)
			if err != nil {
				h.sendError(c, "Failed to get run status")
				return nil
			}

			// Send updated run state
			h.sendEvent(c, "run_state", currentRun)

			// Check steps status
			currentSteps, err := h.runRepo.GetSteps(c.Context(), id)
			if err != nil {
				h.sendError(c, "Failed to get steps status")
				return nil
			}

			// Send updated steps state
			h.sendEvent(c, "steps_state", currentSteps)

			// If run is complete, stop streaming
			if currentRun.Status == "success" || currentRun.Status == "failed" || currentRun.Status == "cancelled" {
				h.sendEvent(c, "complete", fiber.Map{"message": "Run completed"})
				return nil
			}
		}
	}
}

func (h *RunHandler) sendEvent(c *fiber.Ctx, eventType string, data interface{}) error {
	eventData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, eventData)
	_, _ = c.WriteString(msg)
	return nil
}

func (h *RunHandler) sendError(c *fiber.Ctx, message string) error {
	msg := fmt.Sprintf("event: error\ndata: {\"error\": \"%s\"}\n\n", message)
	_, _ = c.WriteString(msg)
	return nil
}

func (h *RunHandler) ListRuns(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	// Parse pagination params
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	// Parse filters
	status := c.Query("status")
	workflowID := c.Query("workflow_id")
	triggeredBy := c.Query("triggered_by")

	runs, total, err := h.runRepo.ListWithPagination(c.Context(), tenantID, page, perPage, status, workflowID, triggeredBy)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list runs"})
	}

	response := models.NewPaginatedResponse(runs, page, perPage, total)
	return c.JSON(response)
}
