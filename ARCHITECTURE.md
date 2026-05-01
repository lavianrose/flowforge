# FlowForge Architecture

## Monorepo

flowforge/

- backend/
- frontend/
- docs/
- infra/

---

## Backend Structure

backend/

- cmd/api
- internal/config
- internal/db
- internal/middleware
- internal/auth
- internal/tenant
- internal/workflow
- internal/execution
- internal/scheduler
- internal/logs
- internal/metrics
- migrations
- tests

---

## Frontend Structure

frontend/

- app
- components
- hooks
- lib
- types

---

## Core Entities

- tenants
- users
- workflows
- workflow_versions
- workflow_runs
- workflow_run_steps
- workflow_logs

All tenant-owned tables require tenant_id.

Use UUID primary keys.

---

## API Routes

POST /api/v1/auth/login
GET /api/v1/workflows
POST /api/v1/workflows
GET /api/v1/workflows/:id
PUT /api/v1/workflows/:id
DELETE /api/v1/workflows/:id
POST /api/v1/workflows/:id/trigger
GET /api/v1/runs/:id
GET /api/v1/runs/:id/stream

---

## Roles

- admin
- editor
- viewer

---

## Workflow Definition

JSON DAG:

{
"nodes": [],
"edges": []
}

Step types:

- http
- delay
- script
- condition

---

## Execution Rules

- Reject cycles
- Topological sort required
- Parallel execution allowed
- Retry with exponential backoff
- Workflow timeout required
- Persist logs and statuses

---

## Realtime

Use SSE first.

Events:

- run_started
- step_started
- step_success
- step_failed
- run_completed

---

## Frontend Pages

/login
/dashboard
/workflows
/workflows/[id]
/runs/[id]

---

## MVP First

Build first:

1. Auth
2. Workflow CRUD
3. Trigger workflow
4. DAG engine
5. SSE monitor
6. Run history

Later:

- Cron
- Webhook
- AI feature
- Better UI
