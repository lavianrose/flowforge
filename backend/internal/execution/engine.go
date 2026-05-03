package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lavianrose/flowforge/internal/dag"
	"github.com/lavianrose/flowforge/internal/models"
	"github.com/lavianrose/flowforge/internal/repository"
)

const (
	defaultNodeRetryCount = 3
	baseRetryDelay        = 500 * time.Millisecond
)

type RunRepository interface {
	Create(ctx context.Context, run *models.WorkflowRun) error
	UpdateStatus(ctx context.Context, id string, status string, errorMsg *string, startedAt, completedAt *time.Time) error
	CreateStep(ctx context.Context, step *models.WorkflowRunStep) error
	UpdateStep(ctx context.Context, step *models.WorkflowRunStep) error
}

type Engine struct {
	runRepo      RunRepository
	workflowRepo *repository.WorkflowRepository
	validator    *dag.Validator
}

func NewEngine(runRepo *repository.RunRepository, workflowRepo *repository.WorkflowRepository) *Engine {
	return &Engine{
		runRepo:      runRepo,
		workflowRepo: workflowRepo,
		validator:    dag.NewValidator(),
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

	// Start execution in background — pass workflow by value to avoid data race
	// on the pointer if the caller mutates it after Execute returns.
	wfCopy := *workflow
	go e.executeWorkflow(context.Background(), &wfCopy, run.ID)

	return run, nil
}

func (e *Engine) executeWorkflow(ctx context.Context, workflow *models.Workflow, runID string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in executeWorkflow: %v\n", r)
			e.failRun(context.Background(), runID, fmt.Sprintf("Execution panicked: %v", r))
		}
	}()

	// Set timeout
	timeout := time.Duration(workflow.TimeoutSecs) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Update status to running
	now := time.Now()
	if err := e.runRepo.UpdateStatus(ctx, runID, "running", nil, &now, nil); err != nil {
		fmt.Printf("Failed to update run status: %v\n", err)
		return
	}

	// Get execution levels for parallel processing
	levels, err := e.validator.GetExecutionLevels(workflow.Definition)
	if err != nil {
		e.failRun(ctx, runID, err.Error())
		return
	}

	// Execute each level
	nodeOutputs := make(map[string]interface{})
	var outputsMu sync.Mutex
	for _, level := range levels {
		// Execute all nodes in this level in parallel
		var wg sync.WaitGroup
		errChan := make(chan error, len(level))

		for _, nodeID := range level {
			wg.Add(1)
			go e.executeNode(ctx, &wg, errChan, workflow, runID, nodeID, nodeOutputs, &outputsMu)
		}

		wg.Wait()
		close(errChan)

		// Check for errors
		var errs []error
		for err := range errChan {
			errs = append(errs, err)
		}

		if len(errs) > 0 {
			e.failRun(ctx, runID, fmt.Sprintf("Execution failed: %v", errs))
			return
		}
	}

	// Mark as completed
	now = time.Now()
	if err := e.runRepo.UpdateStatus(ctx, runID, "success", nil, nil, &now); err != nil {
		fmt.Printf("Failed to update run status: %v\n", err)
	}
}

func (e *Engine) executeNode(ctx context.Context, wg *sync.WaitGroup, errChan chan<- error, workflow *models.Workflow, runID string, nodeID string, outputs map[string]interface{}, outputsMu *sync.Mutex) {
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

	// Snapshot inputs for this node (safe copy of current outputs)
	outputsMu.Lock()
	inputsCopy := make(map[string]interface{})
	for k, v := range outputs {
		inputsCopy[k] = v
	}
	outputsMu.Unlock()

	// Create step
	step := &models.WorkflowRunStep{
		RunID:  runID,
		StepID: nodeID,
		Status: "pending",
		Input:  inputsCopy,
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

	for attempt := 0; attempt < defaultNodeRetryCount; attempt++ {
		output, err := e.executeNodeLogic(ctx, node, inputsCopy)
		if err == nil {
			step.Status = "success"
			step.Output = output
			now = time.Now()
			step.CompletedAt = &now
			if err := e.runRepo.UpdateStep(ctx, step); err != nil {
				errChan <- fmt.Errorf("failed to update step: %w", err)
				return
			}

			outputsMu.Lock()
			outputs[nodeID] = output
			outputsMu.Unlock()
			return
		}

		step.RetryCount = attempt + 1
		step.Error = err.Error()

		if err := e.runRepo.UpdateStep(ctx, step); err != nil {
			errChan <- fmt.Errorf("failed to update step retry metadata: %w", err)
			return
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			step.Status = "failed"
			now = time.Now()
			step.CompletedAt = &now
			e.runRepo.UpdateStep(ctx, step)
			errChan <- err
			return
		}

		if attempt < defaultNodeRetryCount-1 {
			select {
			case <-time.After(time.Duration(1<<attempt) * baseRetryDelay):
			case <-ctx.Done():
				step.Status = "failed"
				now = time.Now()
				step.CompletedAt = &now
				e.runRepo.UpdateStep(ctx, step)
				errChan <- ctx.Err()
				return
			}
			continue
		}

		step.Status = "failed"
		now = time.Now()
		step.CompletedAt = &now
		if err := e.runRepo.UpdateStep(ctx, step); err != nil {
			errChan <- fmt.Errorf("failed to update step: %w", err)
			return
		}
		errChan <- err
		return
	}
}

func (e *Engine) executeNodeLogic(ctx context.Context, node *models.WorkflowNode, inputs map[string]interface{}) (map[string]interface{}, error) {
	switch node.Type {
	case "delay":
		return e.executeDelay(ctx, node)
	case "http":
		return e.executeHTTP(ctx, node, inputs)
	case "script":
		return e.executeScript(ctx, node, inputs)
	case "condition":
		return e.executeCondition(ctx, node, inputs)
	default:
		return nil, fmt.Errorf("unknown node type: %s", node.Type)
	}
}

func (e *Engine) executeDelay(ctx context.Context, node *models.WorkflowNode) (map[string]interface{}, error) {
	seconds := 0
	switch v := node.Config["seconds"].(type) {
	case float64:
		seconds = int(v)
	case string:
		var err error
		seconds, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid seconds value: %v", v)
		}
	default:
		return nil, fmt.Errorf("invalid seconds type in config")
	}

	select {
	case <-time.After(time.Duration(seconds) * time.Second):
		return map[string]interface{}{"message": fmt.Sprintf("Delayed %d seconds", seconds)}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (e *Engine) executeHTTP(ctx context.Context, node *models.WorkflowNode, inputs map[string]interface{}) (map[string]interface{}, error) {
	url, _ := node.Config["url"].(string)
	method, _ := node.Config["method"].(string)

	if url == "" {
		return nil, fmt.Errorf("http node: url is required")
	}
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	var body io.Reader
	if bodyVal, ok := node.Config["body"]; ok && bodyVal != nil {
		switch v := bodyVal.(type) {
		case string:
			// Resolve template variables from inputs
			resolved := e.resolveTemplate(v, inputs)
			body = strings.NewReader(resolved)
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("http node: failed to marshal body: %w", err)
			}
			body = bytes.NewReader(b)
		}
	}

	// Resolve template variables in URL
	url = e.resolveTemplate(url, inputs)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("http node: failed to create request: %w", err)
	}

	// Set headers
	if headers, ok := node.Config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if strVal, ok := v.(string); ok {
				req.Header.Set(k, e.resolveTemplate(strVal, inputs))
			}
		}
	}

	// Set Content-Type default
	if req.Header.Get("Content-Type") == "" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http node: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http node: failed to read response: %w", err)
	}

	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     flattenHeaders(resp.Header),
		"body":        string(respBody),
	}

	// Try to parse body as JSON
	var jsonBody interface{}
	if err := json.Unmarshal(respBody, &jsonBody); err == nil {
		result["json"] = jsonBody
	}

	if resp.StatusCode >= 400 {
		result["error"] = fmt.Sprintf("HTTP request returned status %d", resp.StatusCode)
		return result, fmt.Errorf("http node: request failed with status %d", resp.StatusCode)
	}

	return result, nil
}

func (e *Engine) executeScript(ctx context.Context, node *models.WorkflowNode, inputs map[string]interface{}) (map[string]interface{}, error) {
	code, _ := node.Config["code"].(string)
	if code == "" {
		return nil, fmt.Errorf("script node: code is required")
	}

	// Build a sandboxed evaluation context with inputs available
	result := make(map[string]interface{})

	// Execute simple expressions supported:
	// 1. "transform" mode: expects a "return" statement with JSON-like expression
	// 2. We evaluate simple JSON transforms using Go's text/template style

	// For MVP: support simple JSON path extraction and transformation
	// Pattern: return inputs.node_id.field or return {"key": "value"}
	code = strings.TrimSpace(code)

	// Check if it's a simple return statement with JSON
	if strings.HasPrefix(code, "return ") {
		expr := strings.TrimPrefix(code, "return ")
		expr = strings.TrimSpace(expr)

		// Resolve template variables in the expression
		resolved := e.resolveTemplate(expr, inputs)

		// Try to parse as JSON
		var parsed interface{}
		if err := json.Unmarshal([]byte(resolved), &parsed); err == nil {
			switch v := parsed.(type) {
			case map[string]interface{}:
				return v, nil
			default:
				result["result"] = v
				return result, nil
			}
		}

		// Try as string value
		resolved = strings.Trim(resolved, "\"")
		result["result"] = resolved
		return result, nil
	}

	// For direct JSON transform mode
	var jsonTransform interface{}
	if err := json.Unmarshal([]byte(e.resolveTemplate(code, inputs)), &jsonTransform); err == nil {
		switch v := jsonTransform.(type) {
		case map[string]interface{}:
			return v, nil
		default:
			result["result"] = v
			return result, nil
		}
	}

	// Fallback: return the raw code output with inputs available
	result["output"] = e.resolveTemplate(code, inputs)
	return result, nil
}

func (e *Engine) executeCondition(ctx context.Context, node *models.WorkflowNode, inputs map[string]interface{}) (map[string]interface{}, error) {
	expression, _ := node.Config["expression"].(string)
	if expression == "" {
		return nil, fmt.Errorf("condition node: expression is required")
	}

	// Resolve template variables in expression
	expression = e.resolveTemplate(expression, inputs)

	// Evaluate the expression
	result, err := evaluateExpression(expression)
	if err != nil {
		return nil, fmt.Errorf("condition node: failed to evaluate: %w", err)
	}

	passed, ok := result.(bool)
	if !ok {
		// Try numeric comparison: non-zero = true
		if num, ok := result.(float64); ok {
			passed = num != 0
		} else {
			passed = result != nil
		}
	}

	return map[string]interface{}{
		"passed":     passed,
		"expression": expression,
		"result":     fmt.Sprintf("%v", result),
	}, nil
}

// resolveTemplate replaces {{inputs.node_id}} or {{inputs.node_id.field}} placeholders
// with actual values from the inputs map.
func (e *Engine) resolveTemplate(tmpl string, inputs map[string]interface{}) string {
	result := tmpl

	// Find all {{...}} patterns
	for {
		start := strings.Index(result, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(result, "}}")
		if end == -1 || end <= start {
			break
		}

		path := strings.TrimSpace(result[start+2 : end])

		// Remove "inputs." prefix
		path = strings.TrimPrefix(path, "inputs.")

		var value interface{}
		value = inputs

		// Navigate the path
		parts := strings.Split(path, ".")
		valid := true
		for _, part := range parts {
			if part == "" {
				continue
			}
			m, ok := value.(map[string]interface{})
			if !ok {
				valid = false
				break
			}
			value, ok = m[part]
			if !ok {
				valid = false
				break
			}
		}

		if valid {
			var replacement string
			switch v := value.(type) {
			case string:
				replacement = v
			case nil:
				replacement = ""
			default:
				b, _ := json.Marshal(v)
				replacement = string(b)
			}
			result = result[:start] + replacement + result[end+2:]
		} else {
			// Leave placeholder as-is if path not found
			result = result[:start] + "" + result[end+2:]
		}
	}

	return result
}

// evaluateExpression evaluates simple comparison expressions.
// Supports: ==, !=, >, <, >=, <= operators and basic string/number comparisons.
func evaluateExpression(expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)

	// Handle boolean literals
	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}

	// Try comparison operators (order matters: check >= before >)
	operators := []string{">=", "<=", "!=", "==", ">", "<"}
	for _, op := range operators {
		parts := strings.SplitN(expr, op, 2)
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			leftVal, err := parseValue(left)
			if err != nil {
				return nil, err
			}
			rightVal, err := parseValue(right)
			if err != nil {
				return nil, err
			}

			return compare(leftVal, rightVal, op)
		}
	}

	// If no operator found, try as a value
	return parseValue(expr)
}

func parseValue(s string) (interface{}, error) {
	s = strings.TrimSpace(s)

	// Remove surrounding quotes
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1], nil
	}

	// Try as number
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	}

	// Try as boolean
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}

	// Return as string
	return s, nil
}

func compare(left, right interface{}, op string) (bool, error) {
	// Try numeric comparison
	leftNum, leftIsNum := toFloat64(left)
	rightNum, rightIsNum := toFloat64(right)

	if leftIsNum && rightIsNum {
		switch op {
		case "==":
			return leftNum == rightNum, nil
		case "!=":
			return leftNum != rightNum, nil
		case ">":
			return leftNum > rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		}
	}

	// String comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	switch op {
	case "==":
		return leftStr == rightStr, nil
	case "!=":
		return leftStr != rightStr, nil
	case ">":
		return leftStr > rightStr, nil
	case "<":
		return leftStr < rightStr, nil
	case ">=":
		return leftStr >= rightStr, nil
	case "<=":
		return leftStr <= rightStr, nil
	}

	return false, fmt.Errorf("unsupported operator: %s", op)
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	}
	return 0, false
}

func flattenHeaders(h http.Header) map[string]string {
	m := make(map[string]string)
	for k, vals := range h {
		if len(vals) > 0 {
			m[k] = vals[0]
		}
	}
	return m
}

func (e *Engine) failRun(ctx context.Context, runID string, errorMsg string) {
	now := time.Now()
	if err := e.runRepo.UpdateStatus(ctx, runID, "failed", &errorMsg, nil, &now); err != nil {
		fmt.Printf("Failed to update run status: %v\n", err)
	}
}
