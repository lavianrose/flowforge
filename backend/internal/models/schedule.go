package models

import "time"

type Schedule struct {
	ID            string     `json:"id" db:"id"`
	WorkflowID    string     `json:"workflow_id" db:"workflow_id"`
	TenantID      string     `json:"tenant_id" db:"tenant_id"`
	CronExpression string    `json:"cron_expression" db:"cron_expression"`
	Active        bool       `json:"active" db:"active"`
	NextRunAt     time.Time  `json:"next_run_at" db:"next_run_at"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty" db:"last_run_at"`
	CreatedBy     string     `json:"created_by" db:"created_by"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type Webhook struct {
	ID         string    `json:"id" db:"id"`
	WorkflowID string    `json:"workflow_id" db:"workflow_id"`
	TenantID   string    `json:"tenant_id" db:"tenant_id"`
	Path       string    `json:"path" db:"path"`
	Secret     string    `json:"secret" db:"secret"`
	Active     bool      `json:"active" db:"active"`
	CreatedBy  string    `json:"created_by" db:"created_by"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
