package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lavianrose/flowforge/internal/models"
)

type WorkflowRepository struct {
	conn *pgxpool.Pool
}

func NewWorkflowRepository(conn *pgxpool.Pool) *WorkflowRepository {
	return &WorkflowRepository{conn: conn}
}

func (r *WorkflowRepository) List(ctx context.Context, tenantID string) ([]models.Workflow, error) {
	query := `
		SELECT id, tenant_id, name, description, definition, timeout_seconds, active, created_by, created_at, updated_at
		FROM workflows
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.conn.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []models.Workflow
	for rows.Next() {
		var w models.Workflow
		var defJSON []byte

		err := rows.Scan(
			&w.ID,
			&w.TenantID,
			&w.Name,
			&w.Description,
			&defJSON,
			&w.TimeoutSecs,
			&w.Active,
			&w.CreatedBy,
			&w.CreatedAt,
			&w.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(defJSON, &w.Definition); err != nil {
			return nil, err
		}

		workflows = append(workflows, w)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return workflows, nil
}

func (r *WorkflowRepository) Get(ctx context.Context, id, tenantID string) (*models.Workflow, error) {
	query := `
		SELECT id, tenant_id, name, description, definition, timeout_seconds, active, created_by, created_at, updated_at
		FROM workflows
		WHERE id = $1 AND tenant_id = $2
	`

	var w models.Workflow
	var defJSON []byte

	err := r.conn.QueryRow(ctx, query, id, tenantID).Scan(
		&w.ID,
		&w.TenantID,
		&w.Name,
		&w.Description,
		&defJSON,
		&w.TimeoutSecs,
		&w.Active,
		&w.CreatedBy,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("workflow not found")
		}
		return nil, err
	}

	if err := json.Unmarshal(defJSON, &w.Definition); err != nil {
		return nil, err
	}

	return &w, nil
}

func (r *WorkflowRepository) Create(ctx context.Context, w *models.Workflow) error {
	defJSON, err := json.Marshal(w.Definition)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO workflows (tenant_id, name, description, definition, timeout_seconds, active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err = r.conn.QueryRow(ctx, query,
		w.TenantID,
		w.Name,
		w.Description,
		defJSON,
		w.TimeoutSecs,
		w.Active,
		w.CreatedBy,
	).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)

	return err
}

func (r *WorkflowRepository) Update(ctx context.Context, w *models.Workflow) error {
	defJSON, err := json.Marshal(w.Definition)
	if err != nil {
		return err
	}

	query := `
		UPDATE workflows
		SET name = $1, description = $2, definition = $3, timeout_seconds = $4, active = $5, updated_at = NOW()
		WHERE id = $6 AND tenant_id = $7
		RETURNING updated_at
	`

	err = r.conn.QueryRow(ctx, query,
		w.Name,
		w.Description,
		defJSON,
		w.TimeoutSecs,
		w.Active,
		w.ID,
		w.TenantID,
	).Scan(&w.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("workflow not found")
		}
		return err
	}

	return nil
}

func (r *WorkflowRepository) Delete(ctx context.Context, id, tenantID string) error {
	query := `DELETE FROM workflows WHERE id = $1 AND tenant_id = $2`

	result, err := r.conn.Exec(ctx, query, id, tenantID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("workflow not found")
	}

	return nil
}

func (r *WorkflowRepository) CreateVersion(ctx context.Context, v *models.WorkflowVersion) error {
	defJSON, err := json.Marshal(v.Definition)
	if err != nil {
		return err
	}

	// Get next version number
	var version int
	err = r.conn.QueryRow(ctx,
		`SELECT COALESCE(MAX(version), 0) + 1 FROM workflow_versions WHERE workflow_id = $1`,
		v.WorkflowID,
	).Scan(&version)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO workflow_versions (workflow_id, version, definition, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err = r.conn.QueryRow(ctx, query,
		v.WorkflowID,
		version,
		defJSON,
		v.CreatedBy,
	).Scan(&v.ID, &v.CreatedAt)

	v.Version = version
	return err
}
