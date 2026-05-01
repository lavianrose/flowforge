package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lavianrose/flowforge/internal/models"
)

type RunRepository struct {
	conn *pgxpool.Pool
}

func NewRunRepository(conn *pgxpool.Pool) *RunRepository {
	return &RunRepository{conn: conn}
}

func (r *RunRepository) Create(ctx context.Context, run *models.WorkflowRun) error {
	query := `
		INSERT INTO workflow_runs (workflow_id, tenant_id, status, created_by, triggered_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.conn.QueryRow(ctx, query,
		run.WorkflowID,
		run.TenantID,
		run.Status,
		run.CreatedBy,
		run.TriggeredBy,
	).Scan(&run.ID, &run.CreatedAt)

	return err
}

func (r *RunRepository) Get(ctx context.Context, id, tenantID string) (*models.WorkflowRun, error) {
	query := `
		SELECT id, workflow_id, tenant_id, status, error, started_at, completed_at, created_by, created_at, triggered_by
		FROM workflow_runs
		WHERE id = $1 AND tenant_id = $2
	`

	var run models.WorkflowRun
	err := r.conn.QueryRow(ctx, query, id, tenantID).Scan(
		&run.ID,
		&run.WorkflowID,
		&run.TenantID,
		&run.Status,
		&run.Error,
		&run.StartedAt,
		&run.CompletedAt,
		&run.CreatedBy,
		&run.CreatedAt,
		&run.TriggeredBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("run not found")
		}
		return nil, err
	}

	return &run, nil
}

func (r *RunRepository) List(ctx context.Context, tenantID, workflowID string, limit, offset int) ([]models.WorkflowRun, error) {
	query := `
		SELECT id, workflow_id, tenant_id, status, error, started_at, completed_at, created_by, created_at, triggered_by
		FROM workflow_runs
		WHERE tenant_id = $1
	`

	args := []interface{}{tenantID}
	argIdx := 2

	if workflowID != "" {
		query += fmt.Sprintf(" AND workflow_id = $%d", argIdx)
		args = append(args, workflowID)
		argIdx++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []models.WorkflowRun
	for rows.Next() {
		var run models.WorkflowRun
		err := rows.Scan(
			&run.ID,
			&run.WorkflowID,
			&run.TenantID,
			&run.Status,
			&run.Error,
			&run.StartedAt,
			&run.CompletedAt,
			&run.CreatedBy,
			&run.CreatedAt,
			&run.TriggeredBy,
		)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return runs, nil
}

// ListWithPagination returns runs with pagination and filtering
func (r *RunRepository) ListWithPagination(ctx context.Context, tenantID string, page, perPage int, status, workflowID, triggeredBy string) ([]models.WorkflowRun, int, error) {
	// Build WHERE clause
	whereClause := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argCount := 1

	if status != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	if workflowID != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND workflow_id = $%d", argCount)
		args = append(args, workflowID)
	}

	if triggeredBy != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND triggered_by = $%d", argCount)
		args = append(args, triggeredBy)
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM workflow_runs %s", whereClause)
	var total int
	err := r.conn.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	argCount++
	offset := (page - 1) * perPage
	dataQuery := fmt.Sprintf(`
		SELECT id, workflow_id, tenant_id, status, error, started_at, completed_at, created_by, created_at, triggered_by
		FROM workflow_runs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, perPage, offset)

	rows, err := r.conn.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var runs []models.WorkflowRun
	for rows.Next() {
		var run models.WorkflowRun
		err := rows.Scan(
			&run.ID,
			&run.WorkflowID,
			&run.TenantID,
			&run.Status,
			&run.Error,
			&run.StartedAt,
			&run.CompletedAt,
			&run.CreatedBy,
			&run.CreatedAt,
			&run.TriggeredBy,
		)
		if err != nil {
			return nil, 0, err
		}
		runs = append(runs, run)
	}

	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	return runs, total, nil
}

func (r *RunRepository) UpdateStatus(ctx context.Context, id string, status string, errorMsg *string, startedAt, completedAt **time.Time) error {
	query := `
		UPDATE workflow_runs
		SET status = $1
	`

	args := []interface{}{status}
	argIdx := 2

	if errorMsg != nil {
		query += fmt.Sprintf(", error = $%d", argIdx)
		args = append(args, *errorMsg)
		argIdx++
	}

	if startedAt != nil {
		query += fmt.Sprintf(", started_at = $%d", argIdx)
		args = append(args, *startedAt)
		argIdx++
	}

	if completedAt != nil {
		query += fmt.Sprintf(", completed_at = $%d", argIdx)
		args = append(args, *completedAt)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)

	_, err := r.conn.Exec(ctx, query, args...)
	return err
}

func (r *RunRepository) CreateStep(ctx context.Context, step *models.WorkflowRunStep) error {
	query := `
		INSERT INTO workflow_run_steps (run_id, step_id, status, input, output, error, retry_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err := r.conn.QueryRow(ctx, query,
		step.RunID,
		step.StepID,
		step.Status,
		step.Input,
		step.Output,
		step.Error,
		step.RetryCount,
	).Scan(&step.ID, &step.CreatedAt)

	return err
}

func (r *RunRepository) UpdateStep(ctx context.Context, step *models.WorkflowRunStep) error {
	query := `
		UPDATE workflow_run_steps
		SET status = $1, output = $2, error = $3, retry_count = $4, started_at = $5, completed_at = $6
		WHERE id = $7
	`

	_, err := r.conn.Exec(ctx, query,
		step.Status,
		step.Output,
		step.Error,
		step.RetryCount,
		step.StartedAt,
		step.CompletedAt,
		step.ID,
	)

	return err
}

func (r *RunRepository) GetSteps(ctx context.Context, runID string) ([]models.WorkflowRunStep, error) {
	query := `
		SELECT id, run_id, step_id, status, input, output, error, retry_count, started_at, completed_at, created_at
		FROM workflow_run_steps
		WHERE run_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.conn.Query(ctx, query, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []models.WorkflowRunStep
	for rows.Next() {
		var step models.WorkflowRunStep
		err := rows.Scan(
			&step.ID,
			&step.RunID,
			&step.StepID,
			&step.Status,
			&step.Input,
			&step.Output,
			&step.Error,
			&step.RetryCount,
			&step.StartedAt,
			&step.CompletedAt,
			&step.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return steps, nil
}
