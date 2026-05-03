package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/lavianrose/flowforge/internal/dag"
	"github.com/lavianrose/flowforge/internal/execution"
	"github.com/lavianrose/flowforge/internal/models"
	"github.com/lavianrose/flowforge/internal/repository"
	"github.com/lavianrose/flowforge/internal/validator"
)

type WorkflowHandler struct {
	workflowRepo *repository.WorkflowRepository
	runRepo      *repository.RunRepository
	validator    *dag.Validator
	engine       *execution.Engine
}

func NewWorkflowHandler(workflowRepo *repository.WorkflowRepository, runRepo *repository.RunRepository, engine *execution.Engine) *WorkflowHandler {
	validator := dag.NewValidator()

	return &WorkflowHandler{
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
		validator:    validator,
		engine:       engine,
	}
}

func (h *WorkflowHandler) ListWorkflows(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)

	// Parse and validate pagination params
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	orderBy := c.Query("order_by", "created_at")
	orderDir := c.Query("order_dir", "desc")

	// Validate pagination params
	if !validator.ValidatePage(page) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid page number"})
	}
	if !validator.ValidatePerPage(perPage) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid per_page value (max 100)"})
	}
	if !validator.ValidateOrderBy(orderBy) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid order_by field"})
	}
	if !validator.ValidateOrderDir(orderDir) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid order_dir (must be asc or desc)"})
	}

	// Parse filters
	var activeFilter *bool
	if c.Query("active") != "" {
		active := c.Query("active") == "true"
		activeFilter = &active
	}

	workflows, total, err := h.workflowRepo.ListWithPagination(c.Context(), tenantID, page, perPage, activeFilter, orderBy, orderDir)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list workflows"})
	}

	response := models.NewPaginatedResponse(workflows, page, perPage, total)
	return c.JSON(response)
}

func (h *WorkflowHandler) CreateWorkflow(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(string)

	var req struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Definition  interface{} `json:"definition"`
		TimeoutSecs int         `json:"timeout_seconds"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Sanitize and validate inputs
	v := validator.New()

	// Sanitize
	req.Name = v.SanitizeString(req.Name)
	req.Description = v.SanitizeString(req.Description)

	// Validate
	if err := validator.ValidateWorkflowName(req.Name); len(err) > 0 {
		return c.Status(400).JSON(fiber.Map{"error": err[0]})
	}

	if descErrs := validator.ValidateDescription(req.Description); len(descErrs) > 0 {
		return c.Status(400).JSON(fiber.Map{"error": descErrs[0]})
	}

	if req.Definition == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Definition is required"})
	}

	if req.TimeoutSecs == 0 {
		req.TimeoutSecs = 300 // 5 minutes default
	}

	// Validate timeout (max 1 hour)
	if req.TimeoutSecs < 1 || req.TimeoutSecs > 3600 {
		return c.Status(400).JSON(fiber.Map{"error": "timeout_seconds must be between 1 and 3600"})
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
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to create workflow: %s", err.Error())})
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

func (h *WorkflowHandler) GetWorkflowVersions(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	id := c.Params("id")

	versions, err := h.workflowRepo.GetVersions(c.Context(), id, tenantID)
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(404).JSON(fiber.Map{"error": "Workflow not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get versions"})
	}

	return c.JSON(fiber.Map{"versions": versions})
}

func (h *WorkflowHandler) RollbackWorkflow(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(string)
	userID := c.Locals("user_id").(string)
	id := c.Params("id")
	version := c.Params("version")

	// Parse version number
	var versionNum int
	if _, err := fmt.Sscanf(version, "%d", &versionNum); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid version number"})
	}

	workflow, err := h.workflowRepo.Rollback(c.Context(), id, tenantID, versionNum)
	if err != nil {
		if err.Error() == "workflow not found" || err.Error() == "version not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to rollback workflow"})
	}

	// Create a new version after rollback
	newVersion := &models.WorkflowVersion{
		WorkflowID: id,
		Definition: workflow.Definition,
		CreatedBy:  userID,
	}
	h.workflowRepo.CreateVersion(c.Context(), newVersion)

	return c.JSON(workflow)
}
