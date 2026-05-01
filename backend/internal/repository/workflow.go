package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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

// ListWithPagination returns workflows with pagination and filtering
func (r *WorkflowRepository) ListWithPagination(ctx context.Context, tenantID string, page, perPage int, active *bool, orderBy, orderDir string) ([]models.Workflow, int, error) {
	// Build WHERE clause
	whereClause := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argCount := 1

	if active != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND active = $%d", argCount)
		args = append(args, *active)
	}

	// Build ORDER BY clause
	orderClause := fmt.Sprintf("ORDER BY %s %s", orderBy, orderDir)

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM workflows %s", whereClause)
	var total int
	err := r.conn.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	argCount++
	offset := (page - 1) * perPage
	dataQuery := fmt.Sprintf(`
		SELECT id, tenant_id, name, description, definition, timeout_seconds, active, created_by, created_at, updated_at
		FROM workflows
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argCount, argCount+1)

	args = append(args, perPage, offset)

	rows, err := r.conn.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
		}

		if err := json.Unmarshal(defJSON, &w.Definition); err != nil {
			return nil, 0, err
		}

		workflows = append(workflows, w)
	}

	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	return workflows, total, nil
}

// GetVersions returns all versions of a workflow
func (r *WorkflowRepository) GetVersions(ctx context.Context, workflowID, tenantID string) ([]models.WorkflowVersion, error) {
	// First verify workflow belongs to tenant
	var exists bool
	err := r.conn.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM workflows WHERE id = $1 AND tenant_id = $2)",
		workflowID, tenantID,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("workflow not found")
	}

	query := `
		SELECT id, workflow_id, version, definition, created_by, created_at
		FROM workflow_versions
		WHERE workflow_id = $1
		ORDER BY version DESC
	`

	rows, err := r.conn.Query(ctx, query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []models.WorkflowVersion
	for rows.Next() {
		var v models.WorkflowVersion
		var defJSON []byte

		err := rows.Scan(
			&v.ID,
			&v.WorkflowID,
			&v.Version,
			&defJSON,
			&v.CreatedBy,
			&v.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(defJSON, &v.Definition); err != nil {
			return nil, err
		}

		versions = append(versions, v)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return versions, nil
}

// Rollback restores a workflow to a specific version
func (r *WorkflowRepository) Rollback(ctx context.Context, workflowID, tenantID string, version int) (*models.Workflow, error) {
	// Get version
	query := `
		SELECT definition
		FROM workflow_versions
		WHERE workflow_id = $1 AND version = $2
	`

	var defJSON []byte
	err := r.conn.QueryRow(ctx, query, workflowID, version).Scan(&defJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("version not found")
		}
		return nil, err
	}

	// Get current workflow
	workflow, err := r.Get(ctx, workflowID, tenantID)
	if err != nil {
		return nil, err
	}

	// Update workflow with version definition
	var definition models.WorkflowDef
	if err := json.Unmarshal(defJSON, &definition); err != nil {
		return nil, err
	}
	workflow.Definition = definition

	updateQuery := `
		UPDATE workflows
		SET definition = $1, updated_at = NOW()
		WHERE id = $2 AND tenant_id = $3
		RETURNING updated_at
	`

	err = r.conn.QueryRow(ctx, updateQuery, defJSON, workflowID, tenantID).Scan(&workflow.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}
