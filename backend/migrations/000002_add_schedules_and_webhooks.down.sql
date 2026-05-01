-- Drop triggers
DROP TRIGGER IF EXISTS update_schedules_updated_at ON schedules;
DROP TRIGGER IF EXISTS update_webhooks_updated_at ON webhooks;

-- Drop indexes
DROP INDEX IF EXISTS idx_schedules_active;
DROP INDEX IF EXISTS idx_schedules_next_run_at;
DROP INDEX IF EXISTS idx_schedules_tenant_id;
DROP INDEX IF EXISTS idx_schedules_workflow_id;

DROP INDEX IF EXISTS idx_webhooks_active;
DROP INDEX IF EXISTS idx_webhooks_path;
DROP INDEX IF EXISTS idx_webhooks_tenant_id;
DROP INDEX IF EXISTS idx_webhooks_workflow_id;

-- Drop tables
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS schedules;
