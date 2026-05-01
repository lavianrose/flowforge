package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lavianrose/flowforge/internal/models"
)

type ScheduleRepository struct {
	conn *pgxpool.Pool
}

func NewScheduleRepository(conn *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{conn: conn}
}

func (r *ScheduleRepository) Create(ctx context.Context, s *models.Schedule) error {
	query := `
		INSERT INTO schedules (workflow_id, tenant_id, cron_expression, active, next_run_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.conn.QueryRow(ctx, query,
		s.WorkflowID,
		s.TenantID,
		s.CronExpression,
		s.Active,
		s.NextRunAt,
		s.CreatedBy,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	return err
}

func (r *ScheduleRepository) Get(ctx context.Context, id, tenantID string) (*models.Schedule, error) {
	query := `
		SELECT id, workflow_id, tenant_id, cron_expression, active, next_run_at, last_run_at, created_by, created_at, updated_at
		FROM schedules
		WHERE id = $1 AND tenant_id = $2
	`

	var s models.Schedule
	err := r.conn.QueryRow(ctx, query, id, tenantID).Scan(
		&s.ID,
		&s.WorkflowID,
		&s.TenantID,
		&s.CronExpression,
		&s.Active,
		&s.NextRunAt,
		&s.LastRunAt,
		&s.CreatedBy,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("schedule not found")
		}
		return nil, err
	}

	return &s, nil
}

func (r *ScheduleRepository) List(ctx context.Context, tenantID string) ([]models.Schedule, error) {
	query := `
		SELECT id, workflow_id, tenant_id, cron_expression, active, next_run_at, last_run_at, created_by, created_at, updated_at
		FROM schedules
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.conn.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var s models.Schedule
		err := rows.Scan(
			&s.ID,
			&s.WorkflowID,
			&s.TenantID,
			&s.CronExpression,
			&s.Active,
			&s.NextRunAt,
			&s.LastRunAt,
			&s.CreatedBy,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return schedules, nil
}

func (r *ScheduleRepository) Update(ctx context.Context, s *models.Schedule) error {
	query := `
		UPDATE schedules
		SET cron_expression = $1, active = $2, next_run_at = $3, updated_at = NOW()
		WHERE id = $4 AND tenant_id = $5
		RETURNING updated_at
	`

	err := r.conn.QueryRow(ctx, query,
		s.CronExpression,
		s.Active,
		s.NextRunAt,
		s.ID,
		s.TenantID,
	).Scan(&s.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("schedule not found")
		}
		return err
	}

	return nil
}

func (r *ScheduleRepository) Delete(ctx context.Context, id, tenantID string) error {
	query := `DELETE FROM schedules WHERE id = $1 AND tenant_id = $2`

	result, err := r.conn.Exec(ctx, query, id, tenantID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("schedule not found")
	}

	return nil
}

func (r *ScheduleRepository) UpdateNextRun(ctx context.Context, id string, nextRun time.Time) error {
	query := `UPDATE schedules SET next_run_at = $1 WHERE id = $2`
	_, err := r.conn.Exec(ctx, query, nextRun, id)
	return err
}

func (r *ScheduleRepository) UpdateLastRun(ctx context.Context, id string, lastRun time.Time) error {
	query := `UPDATE schedules SET last_run_at = $1 WHERE id = $2`
	_, err := r.conn.Exec(ctx, query, lastRun, id)
	return err
}

func (r *ScheduleRepository) GetDueSchedules(ctx context.Context) ([]models.Schedule, error) {
	query := `
		SELECT id, workflow_id, tenant_id, cron_expression, active, next_run_at, last_run_at, created_by, created_at, updated_at
		FROM schedules
		WHERE active = true AND next_run_at <= NOW()
		ORDER BY next_run_at ASC
	`

	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var s models.Schedule
		err := rows.Scan(
			&s.ID,
			&s.WorkflowID,
			&s.TenantID,
			&s.CronExpression,
			&s.Active,
			&s.NextRunAt,
			&s.LastRunAt,
			&s.CreatedBy,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return schedules, nil
}
