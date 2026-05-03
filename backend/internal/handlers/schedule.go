package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	cronutil "github.com/lavianrose/flowforge/internal/cron"
	"github.com/lavianrose/flowforge/internal/execution"
	"github.com/lavianrose/flowforge/internal/models"
	"github.com/lavianrose/flowforge/internal/repository"
	"github.com/lavianrose/flowforge/internal/validator"
)

type ScheduleHandler struct {
	scheduleRepo  *repository.ScheduleRepository
	workflowRepo  *repository.WorkflowRepository
	runRepo       *repository.RunRepository
	engine        *execution.Engine
}

func NewScheduleHandler(
	scheduleRepo *repository.ScheduleRepository,
	workflowRepo *repository.WorkflowRepository,
	runRepo *repository.RunRepository,
) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleRepo: scheduleRepo,
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
	}
}

func (h *ScheduleHandler) CreateSchedule(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(string)

	var req struct {
		WorkflowID    string `json:"workflow_id"`
		CronExpression string `json:"cron_expression"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate
	v := validator.New()
	req.CronExpression = v.SanitizeString(req.CronExpression)
	v.Required("cron_expression", req.CronExpression)
	v.Cron("cron_expression", req.CronExpression)

	if v.HasErrors() {
		return c.Status(400).JSON(fiber.Map{"error": v.ErrorMap()})
	}

	// Verify workflow exists and belongs to tenant
	_, err := h.workflowRepo.Get(c.Context(), req.WorkflowID, tenantID)
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Workflow not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get workflow"})
	}

	// Calculate next run time based on cron expression
	nextRun, err := cronutil.NextRun(req.CronExpression, time.Now())
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid cron expression"})
	}

	schedule := &models.Schedule{
		WorkflowID:    req.WorkflowID,
		TenantID:      tenantID,
		CronExpression: req.CronExpression,
		Active:        true,
		NextRunAt:     nextRun,
		CreatedBy:     userID,
	}

	if err := h.scheduleRepo.Create(c.Context(), schedule); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create schedule"})
	}

	return c.Status(201).JSON(schedule)
}

func (h *ScheduleHandler) ListSchedules(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	schedules, err := h.scheduleRepo.List(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list schedules"})
	}

	return c.JSON(fiber.Map{"schedules": schedules})
}

func (h *ScheduleHandler) DeleteSchedule(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	id := c.Params("id")

	if err := h.scheduleRepo.Delete(c.Context(), id, tenantID); err != nil {
		if err.Error() == "schedule not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Schedule not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete schedule"})
	}

	return c.Status(204).Send(nil)
}

type WebhookHandler struct {
	webhookRepo  *repository.WebhookRepository
	workflowRepo *repository.WorkflowRepository
	runRepo      *repository.RunRepository
	engine       *execution.Engine
}

func NewWebhookHandler(
	webhookRepo *repository.WebhookRepository,
	workflowRepo *repository.WorkflowRepository,
	runRepo *repository.RunRepository,
	engine *execution.Engine,
) *WebhookHandler {
	return &WebhookHandler{
		webhookRepo:  webhookRepo,
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
		engine:       engine,
	}
}

func (h *WebhookHandler) CreateWebhook(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(string)

	var req struct {
		WorkflowID string `json:"workflow_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.WorkflowID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "workflow_id is required"})
	}

	// Verify workflow exists and belongs to tenant
	_, err := h.workflowRepo.Get(c.Context(), req.WorkflowID, tenantID)
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Workflow not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get workflow"})
	}

	// Generate unique path and secret
	path := generatePath()
	secret := generateSecret()

	webhook := &models.Webhook{
		WorkflowID: req.WorkflowID,
		TenantID:   tenantID,
		Path:       path,
		Secret:     secret,
		Active:     true,
		CreatedBy:  userID,
	}

	if err := h.webhookRepo.Create(c.Context(), webhook); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create webhook"})
	}

	return c.Status(201).JSON(webhook)
}

func (h *WebhookHandler) ListWebhooks(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	webhooks, err := h.webhookRepo.List(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list webhooks"})
	}

	return c.JSON(fiber.Map{"webhooks": webhooks})
}

func (h *WebhookHandler) DeleteWebhook(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	id := c.Params("id")

	if err := h.webhookRepo.Delete(c.Context(), id, tenantID); err != nil {
		if err.Error() == "webhook not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Webhook not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete webhook"})
	}

	return c.Status(204).Send(nil)
}

func (h *WebhookHandler) TriggerWebhook(c *fiber.Ctx) error {
	path := c.Params("path")

	// Get webhook by path
	webhook, err := h.webhookRepo.GetByPath(c.Context(), path)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Webhook not found"})
	}

	// Verify signature (simple for now - X-Webhook-Secret header)
	secret := c.Get("X-Webhook-Secret")
	if secret != webhook.Secret {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid webhook secret"})
	}

	// Trigger workflow
	run, err := h.engine.Execute(c.Context(), webhook.WorkflowID, webhook.TenantID, "webhook", nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(202).JSON(run)
}

func generatePath() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "/webhooks/" + hex.EncodeToString(b)
}

func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
