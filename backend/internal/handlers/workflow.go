package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/lavianrose/flowforge/internal/dag"
	"github.com/lavianrose/flowforge/internal/execution"
	"github.com/lavianrose/flowforge/internal/models"
	"github.com/lavianrose/flowforge/internal/repository"
)

type WorkflowHandler struct {
	workflowRepo *repository.WorkflowRepository
	runRepo      *repository.RunRepository
	validator    *dag.Validator
	engine       *execution.Engine
}

func NewWorkflowHandler(workflowRepo *repository.WorkflowRepository, runRepo *repository.RunRepository) *WorkflowHandler {
	validator := dag.NewValidator()
	engine := execution.NewEngine(runRepo, workflowRepo)

	return &WorkflowHandler{
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
		validator:    validator,
		engine:       engine,
	}
}

func (h *WorkflowHandler) ListWorkflows(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	workflows, err := h.workflowRepo.List(c.Context(), tenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list workflows"})
	}

	return c.JSON(fiber.Map{"workflows": workflows})
}

func (h *WorkflowHandler) CreateWorkflow(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(string)

	var req struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Definition  interface{}            `json:"definition"`
		TimeoutSecs int                    `json:"timeout_seconds"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name is required"})
	}

	if req.Definition == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Definition is required"})
	}

	if req.TimeoutSecs == 0 {
		req.TimeoutSecs = 300 // 5 minutes default
	}

	// Convert definition to WorkflowDef
	defBytes, err := json.Marshal(req.Definition)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid definition format"})
	}

	var definition models.WorkflowDef
	if err := json.Unmarshal(defBytes, &definition); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid definition structure"})
	}

	// Validate DAG structure (cycle detection, etc.)
	if err := h.validator.Validate(definition); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	workflow := &models.Workflow{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Definition:  definition,
		TimeoutSecs: req.TimeoutSecs,
		Active:      true,
		CreatedBy:   userID,
	}

	if err := h.workflowRepo.Create(c.Context(), workflow); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create workflow"})
	}

	return c.Status(201).JSON(workflow)
}

func (h *WorkflowHandler) GetWorkflow(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	id := c.Params("id")

	workflow, err := h.workflowRepo.Get(c.Context(), id, tenantID)
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Workflow not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get workflow"})
	}

	return c.JSON(workflow)
}

func (h *WorkflowHandler) UpdateWorkflow(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(string)
	id := c.Params("id")

	var req struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Definition  interface{} `json:"definition"`
		TimeoutSecs int         `json:"timeout_seconds"`
		Active      *bool       `json:"active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	workflow, err := h.workflowRepo.Get(c.Context(), id, tenantID)
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Workflow not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get workflow"})
	}

	if req.Name != "" {
		workflow.Name = req.Name
	}
	if req.Description != "" {
		workflow.Description = req.Description
	}
	if req.Definition != nil {
		// Convert definition to WorkflowDef
		defBytes, err := json.Marshal(req.Definition)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid definition format"})
		}

		var definition models.WorkflowDef
		if err := json.Unmarshal(defBytes, &definition); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid definition structure"})
		}

		// Validate DAG structure
		if err := h.validator.Validate(definition); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		workflow.Definition = definition
	}
	if req.TimeoutSecs > 0 {
		workflow.TimeoutSecs = req.TimeoutSecs
	}
	if req.Active != nil {
		workflow.Active = *req.Active
	}

	if err := h.workflowRepo.Update(c.Context(), workflow); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update workflow"})
	}

	// Create version
	version := &models.WorkflowVersion{
		WorkflowID: id,
		Definition: workflow.Definition,
		CreatedBy:  userID,
	}
	h.workflowRepo.CreateVersion(c.Context(), version)

	return c.JSON(workflow)
}

func (h *WorkflowHandler) DeleteWorkflow(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	id := c.Params("id")

	if err := h.workflowRepo.Delete(c.Context(), id, tenantID); err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Workflow not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete workflow"})
	}

	return c.Status(204).Send(nil)
}

func (h *WorkflowHandler) TriggerWorkflow(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(string)
	id := c.Params("id")

	run, err := h.engine.Execute(c.Context(), id, tenantID, "manual", &userID)
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Workflow not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(202).JSON(run)
}
