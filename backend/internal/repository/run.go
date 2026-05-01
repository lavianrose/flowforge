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

// GetHealthStats returns aggregated statistics for the dashboard
func (r *RunRepository) GetHealthStats(ctx context.Context, tenantID string) (*models.HealthStats, error) {
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	// Get active runs
	var activeRuns int64
	err := r.conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM workflow_runs WHERE tenant_id = $1 AND status IN ('pending', 'running')`,
		tenantID,
	).Scan(&activeRuns)
	if err != nil {
		return nil, err
	}

	// Get total runs in last 24h
	var totalRuns24h int64
	err = r.conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM workflow_runs WHERE tenant_id = $1 AND created_at >= $2`,
		tenantID, twentyFourHoursAgo,
	).Scan(&totalRuns24h)
	if err != nil {
		return nil, err
	}

	// Get success/failed runs in last 24h
	var successRuns24h, failedRuns24h int64
	err = r.conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM workflow_runs WHERE tenant_id = $1 AND created_at >= $2 AND status = 'success'`,
		tenantID, twentyFourHoursAgo,
	).Scan(&successRuns24h)
	if err != nil {
		return nil, err
	}

	err = r.conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM workflow_runs WHERE tenant_id = $1 AND created_at >= $2 AND status = 'failed'`,
		tenantID, twentyFourHoursAgo,
	).Scan(&failedRuns24h)
	if err != nil {
		return nil, err
	}

	// Calculate rates
	var successRate, failureRate float64
	if totalRuns24h > 0 {
		successRate = float64(successRuns24h) / float64(totalRuns24h) * 100
		failureRate = float64(failedRuns24h) / float64(totalRuns24h) * 100
	}

	// Get average duration for completed runs
	var avgDuration float64
	err = r.conn.QueryRow(ctx,
		`SELECT AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) FROM workflow_runs WHERE tenant_id = $1 AND created_at >= $2 AND status IN ('success', 'failed') AND completed_at IS NOT NULL AND started_at IS NOT NULL`,
		tenantID, twentyFourHoursAgo,
	).Scan(&avgDuration)
	if err != nil {
		avgDuration = 0
	}

	// Get hourly stats for the last 24h
	hourlyQuery := `
		SELECT
			EXTRACT(HOUR FROM created_at) AS hour,
			COUNT(*) AS total_runs,
			COUNT(*) FILTER (WHERE status = 'success') AS success_runs,
			COUNT(*) FILTER (WHERE status = 'failed') AS failed_runs,
			COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) FILTER (WHERE status IN ('success', 'failed') AND completed_at IS NOT NULL AND started_at IS NOT NULL), 0) AS avg_duration
		FROM workflow_runs
		WHERE tenant_id = $1 AND created_at >= $2
		GROUP BY EXTRACT(HOUR FROM created_at)
		ORDER BY hour
	`

	rows, err := r.conn.Query(ctx, hourlyQuery, tenantID, twentyFourHoursAgo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hourlyStats []models.HourlyStats
	for rows.Next() {
		var stat models.HourlyStats
		err := rows.Scan(
			&stat.Hour,
			&stat.TotalRuns,
			&stat.SuccessRuns,
			&stat.FailedRuns,
			&stat.AvgDuration,
		)
		if err != nil {
			return nil, err
		}
		hourlyStats = append(hourlyStats, stat)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return &models.HealthStats{
		ActiveRuns:     activeRuns,
		SuccessRate:    successRate,
		FailureRate:    failureRate,
		AvgDuration:    avgDuration,
		TotalRuns24h:   totalRuns24h,
		SuccessRuns24h: successRuns24h,
		FailedRuns24h:  failedRuns24h,
		HourlyStats:    hourlyStats,
	}, nil
}
