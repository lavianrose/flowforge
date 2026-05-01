package execution

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lavianrose/flowforge/internal/dag"
	"github.com/lavianrose/flowforge/internal/models"
	"github.com/lavianrose/flowforge/internal/repository"
)

type Engine struct {
	runRepo    *repository.RunRepository
	workflowRepo *repository.WorkflowRepository
	validator  *dag.Validator
}

func NewEngine(runRepo *repository.RunRepository, workflowRepo *repository.WorkflowRepository) *Engine {
	return &Engine{
		runRepo:    runRepo,
		workflowRepo: workflowRepo,
		validator:  dag.NewValidator(),
	}
}

func (e *Engine) Execute(ctx context.Context, workflowID, tenantID string, triggeredBy string, createdBy *string) (*models.WorkflowRun, error) {
	// Get workflow
	workflow, err := e.workflowRepo.Get(ctx, workflowID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	if !workflow.Active {
		return nil, fmt.Errorf("workflow is not active")
	}

	// Create run
	run := &models.WorkflowRun{
		WorkflowID:  workflowID,
		TenantID:    tenantID,
		Status:      "pending",
		TriggeredBy: triggeredBy,
		CreatedBy:   createdBy,
	}

	if err := e.runRepo.Create(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	// Start execution in background
	go e.executeWorkflow(context.Background(), workflow, run)

	return run, nil
}

func (e *Engine) executeWorkflow(ctx context.Context, workflow *models.Workflow, run *models.WorkflowRun) {
	// Set timeout
	timeout := time.Duration(workflow.TimeoutSecs) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Update status to running
	now := time.Now()
	run.Status = "running"
	run.StartedAt = &now
	if err := e.runRepo.UpdateStatus(ctx, run.ID, run.Status, nil, &run.StartedAt, nil); err != nil {
		fmt.Printf("Failed to update run status: %v\n", err)
		return
	}

	// Get execution levels for parallel processing
	levels, err := e.validator.GetExecutionLevels(workflow.Definition)
	if err != nil {
		e.failRun(ctx, run, err.Error())
		return
	}

	// Execute each level
	nodeOutputs := make(map[string]interface{})
	for _, level := range levels {
		// Execute all nodes in this level in parallel
		var wg sync.WaitGroup
		errChan := make(chan error, len(level))

		for _, nodeID := range level {
			wg.Add(1)
			go e.executeNode(ctx, &wg, errChan, workflow, run, nodeID, nodeOutputs)
		}

		wg.Wait()
		close(errChan)

		// Check for errors
		var errs []error
		for err := range errChan {
			errs = append(errs, err)
		}

		if len(errs) > 0 {
			e.failRun(ctx, run, fmt.Sprintf("Execution failed: %v", errs))
			return
		}
	}

	// Mark as completed
	now = time.Now()
	run.Status = "success"
	run.CompletedAt = &now
	if err := e.runRepo.UpdateStatus(ctx, run.ID, run.Status, nil, nil, &run.CompletedAt); err != nil {
		fmt.Printf("Failed to update run status: %v\n", err)
	}
}

func (e *Engine) executeNode(ctx context.Context, wg *sync.WaitGroup, errChan chan<- error, workflow *models.Workflow, run *models.WorkflowRun, nodeID string, outputs map[string]interface{}) {
	defer wg.Done()

	// Find node
	var node *models.WorkflowNode
	for i := range workflow.Definition.Nodes {
		if workflow.Definition.Nodes[i].ID == nodeID {
			node = &workflow.Definition.Nodes[i]
			break
		}
	}

	if node == nil {
		errChan <- fmt.Errorf("node not found: %s", nodeID)
		return
	}

	// Create step
	step := &models.WorkflowRunStep{
		RunID:  run.ID,
		StepID: nodeID,
		Status: "pending",
		Input:  outputs,
	}

	if err := e.runRepo.CreateStep(ctx, step); err != nil {
		errChan <- fmt.Errorf("failed to create step: %w", err)
		return
	}

	// Update step to running
	now := time.Now()
	step.Status = "running"
	step.StartedAt = &now
	if err := e.runRepo.UpdateStep(ctx, step); err != nil {
		errChan <- fmt.Errorf("failed to update step: %w", err)
		return
	}

	// Execute based on type
	output, err := e.executeNodeLogic(ctx, node, outputs)
	if err != nil {
		step.Status = "failed"
		step.Error = err.Error()
		now = time.Now()
		step.CompletedAt = &now
		e.runRepo.UpdateStep(ctx, step)
		errChan <- err
		return
	}

	// Update step to success
	step.Status = "success"
	step.Output = output
	now = time.Now()
	step.CompletedAt = &now
	if err := e.runRepo.UpdateStep(ctx, step); err != nil {
		errChan <- fmt.Errorf("failed to update step: %w", err)
		return
	}

	// Store output for next nodes
	outputs[nodeID] = output
}

func (e *Engine) executeNodeLogic(ctx context.Context, node *models.WorkflowNode, inputs map[string]interface{}) (map[string]interface{}, error) {
	switch node.Type {
	case "delay":
		seconds := int(node.Config["seconds"].(float64))
		time.Sleep(time.Duration(seconds) * time.Second)
		return map[string]interface{}{"message": fmt.Sprintf("Delayed %d seconds", seconds)}, nil

	case "http":
		// TODO: Implement HTTP request
		return map[string]interface{}{"message": "HTTP execution not implemented"}, nil

	case "script":
		// TODO: Implement script execution
		return map[string]interface{}{"message": "Script execution not implemented"}, nil

	case "condition":
		// TODO: Implement condition evaluation
		return map[string]interface{}{"message": "Condition evaluation not implemented"}, nil

	default:
		return nil, fmt.Errorf("unknown node type: %s", node.Type)
	}
}

func (e *Engine) failRun(ctx context.Context, run *models.WorkflowRun, errorMsg string) {
	now := time.Now()
	run.Status = "failed"
	run.Error = errorMsg
	run.CompletedAt = &now
	if err := e.runRepo.UpdateStatus(ctx, run.ID, run.Status, &run.Error, nil, &run.CompletedAt); err != nil {
		fmt.Printf("Failed to update run status: %v\n", err)
	}
}
