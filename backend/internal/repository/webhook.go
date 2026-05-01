package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lavianrose/flowforge/internal/models"
)

type WebhookRepository struct {
	conn *pgxpool.Pool
}

func NewWebhookRepository(conn *pgxpool.Pool) *WebhookRepository {
	return &WebhookRepository{conn: conn}
}

func (r *WebhookRepository) Create(ctx context.Context, w *models.Webhook) error {
	query := `
		INSERT INTO webhooks (workflow_id, tenant_id, path, secret, active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.conn.QueryRow(ctx, query,
		w.WorkflowID,
		w.TenantID,
		w.Path,
		w.Secret,
		w.Active,
		w.CreatedBy,
	).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)

	return err
}

func (r *WebhookRepository) Get(ctx context.Context, id, tenantID string) (*models.Webhook, error) {
	query := `
		SELECT id, workflow_id, tenant_id, path, secret, active, created_by, created_at, updated_at
		FROM webhooks
		WHERE id = $1 AND tenant_id = $2
	`

	var w models.Webhook
	err := r.conn.QueryRow(ctx, query, id, tenantID).Scan(
		&w.ID,
		&w.WorkflowID,
		&w.TenantID,
		&w.Path,
		&w.Secret,
		&w.Active,
		&w.CreatedBy,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("webhook not found")
		}
		return nil, err
	}

	return &w, nil
}

func (r *WebhookRepository) GetByPath(ctx context.Context, path string) (*models.Webhook, error) {
	query := `
		SELECT id, workflow_id, tenant_id, path, secret, active, created_by, created_at, updated_at
		FROM webhooks
		WHERE path = $1 AND active = true
	`

	var w models.Webhook
	err := r.conn.QueryRow(ctx, query, path).Scan(
		&w.ID,
		&w.WorkflowID,
		&w.TenantID,
		&w.Path,
		&w.Secret,
		&w.Active,
		&w.CreatedBy,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("webhook not found")
		}
		return nil, err
	}

	return &w, nil
}

func (r *WebhookRepository) List(ctx context.Context, tenantID string) ([]models.Webhook, error) {
	query := `
		SELECT id, workflow_id, tenant_id, path, secret, active, created_by, created_at, updated_at
		FROM webhooks
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.conn.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []models.Webhook
	for rows.Next() {
		var w models.Webhook
		err := rows.Scan(
			&w.ID,
			&w.WorkflowID,
			&w.TenantID,
			&w.Path,
			&w.Secret,
			&w.Active,
			&w.CreatedBy,
			&w.CreatedAt,
			&w.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, w)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return webhooks, nil
}

func (r *WebhookRepository) Delete(ctx context.Context, id, tenantID string) error {
	query := `DELETE FROM webhooks WHERE id = $1 AND tenant_id = $2`

	result, err := r.conn.Exec(ctx, query, id, tenantID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("webhook not found")
	}

	return nil
}
