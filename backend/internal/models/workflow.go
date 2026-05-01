package models

import (
	"time"
)

type Workflow struct {
	ID            string          `json:"id" db:"id"`
	TenantID      string          `json:"tenant_id" db:"tenant_id"`
	Name          string          `json:"name" db:"name"`
	Description   string          `json:"description" db:"description"`
	Definition    WorkflowDef     `json:"definition" db:"definition"`
	TimeoutSecs   int             `json:"timeout_seconds" db:"timeout_seconds"`
	Active        bool            `json:"active" db:"active"`
	CreatedBy     string          `json:"created_by" db:"created_by"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

type WorkflowDef struct {
	Nodes []WorkflowNode `json:"nodes"`
	Edges []WorkflowEdge `json:"edges"`
}

type WorkflowNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // http, delay, script, condition
	Name     string                 `json:"name"`
	Config   map[string]interface{} `json:"config"`
	Position map[string]float64     `json:"position"` // {x, y}
}

type WorkflowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type WorkflowVersion struct {
	ID         string       `json:"id" db:"id"`
	WorkflowID string       `json:"workflow_id" db:"workflow_id"`
	Version    int          `json:"version" db:"version"`
	Definition WorkflowDef  `json:"definition" db:"definition"`
	CreatedBy  string       `json:"created_by" db:"created_by"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
}

type WorkflowRun struct {
	ID          string     `json:"id" db:"id"`
	WorkflowID  string     `json:"workflow_id" db:"workflow_id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	Status      string     `json:"status" db:"status"` // pending, running, success, failed, cancelled
	Error       string     `json:"error,omitempty" db:"error"`
	StartedAt   *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedBy   *string    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	TriggeredBy string     `json:"triggered_by" db:"triggered_by"`
}

type WorkflowRunStep struct {
	ID         string                 `json:"id" db:"id"`
	RunID      string                 `json:"run_id" db:"run_id"`
	StepID     string                 `json:"step_id" db:"step_id"`
	Status     string                 `json:"status" db:"status"` // pending, running, success, failed, skipped
	Input      map[string]interface{} `json:"input,omitempty" db:"input"`
	Output     map[string]interface{} `json:"output,omitempty" db:"output"`
	Error      string                 `json:"error,omitempty" db:"error"`
	RetryCount int                    `json:"retry_count" db:"retry_count"`
	StartedAt  *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time            `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

type WorkflowLog struct {
	ID       string                 `json:"id" db:"id"`
	RunID    string                 `json:"run_id" db:"run_id"`
	StepID   string                 `json:"step_id,omitempty" db:"step_id"`
	Level    string                 `json:"level" db:"level"` // debug, info, warn, error
	Message  string                 `json:"message" db:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt time.Time             `json:"created_at" db:"created_at"`
}
