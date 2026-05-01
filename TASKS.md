# FlowForge Tasks

## Global Rule

Always complete highest priority unfinished item first.

---

- [x] Init backend Golang Fiber
- [x] Env config
- [x] PostgreSQL connection
- [x] Redis connection
- [x] Base router
- [x] Migration system
- [x] Create core tables
- [x] Login API
- [x] JWT middleware
- [x] Role middleware
- [x] Workflow CRUD API
- [x] Workflow versioning
- [x] DAG validator
- [x] Cycle detection
- [x] Topological sort
- [x] Trigger workflow
- [x] Execution worker
- [x] Retry logic
- [x] Timeout logic
- [x] Init Next.js app
- [x] Login page
- [x] Dashboard layout
- [x] Workflow list page
- [x] Workflow detail page
- [x] React Flow DAG viewer
- [x] SSE live monitor
- [x] Run history page
- [x] Dockerfile backend
- [x] Dockerfile frontend
- [x] docker-compose.yml
- [ ] GitHub Actions CI
- [ ] Unit tests
- [ ] Integration tests
- [x] README.md
- [ ] ARCHITECTURE.md final review
<!-- - [ ] AI failure analyzer -->

---

## Current Focus

✅ Backend core completed
✅ Frontend foundation completed
✅ React Flow DAG builder completed
✅ SSE live monitoring completed
✅ Docker deployment completed
✅ Documentation completed

**MVP COMPLETE!** 🎉

---

## Priority: Missing Core Backend Features

### High Priority - Must Have

- [ ] Add workflow version rollback endpoint
  - [ ] GET /api/v1/workflows/:id/versions (list all versions)
  - [ ] POST /api/v1/workflows/:id/rollback/:version (restore specific version)
  - [ ] Update repository with GetVersions and Rollback methods

- [ ] Add scheduled/cron triggering
  - [ ] Create schedules table (id, workflow_id, tenant_id, cron_expression, active, next_run_at)
  - [ ] Add cron parser library (github.com/robfig/cron/v3)
  - [ ] Create scheduler service that checks and triggers scheduled workflows
  - [ ] Add API endpoints (POST /api/v1/workflows/:id/schedule, GET /api/v1/schedules, DELETE /api/v1/schedules/:id)

- [ ] Add webhook triggering
  - [ ] Create webhooks table (id, workflow_id, tenant_id, path, secret, active)
  - [ ] Generate unique webhook URLs (e.g., /webhooks/{uuid})
  - [ ] Add webhook handler with signature verification
  - [ ] Add API endpoints (POST /api/v1/workflows/:id/webhook, GET /api/v1/webhooks, DELETE /api/v1/webhooks/:id)

- [ ] Add pagination to all list endpoints
  - [ ] Update ListWorkflows with limit/offset/cursor pagination
  - [ ] Update ListRuns with limit/offset/cursor pagination
  - [ ] Add response metadata (total, page, per_page, total_pages)

- [ ] Add filtering to all list endpoints
  - [ ] Workflows: filter by active, created_by, date range
  - [ ] Runs: filter by status, workflow_id, triggered_by, date range
  - [ ] Add query parameter parsing and validation

- [ ] Add rate limiting middleware
  - [ ] Redis-based rate limiter (sliding window)
  - [ ] Configurable limits per endpoint/role
  - [ ] Add rate limit headers to responses

- [ ] Enforce role-based access control
  - [ ] Apply Role middleware to all endpoints
  - [ ] Viewer: read-only access
  - [ ] Editor: create, update, trigger workflows
  - [ ] Admin: full access including delete

- [ ] Add comprehensive input validation
  - [ ] Add field length limits (name: 255, description: 5000)
  - [ ] Add format validation (email, UUID, cron expression)
  - [ ] Add input sanitization (trim spaces, escape HTML)
  - [ ] Add custom validator middleware

### Medium Priority

- [ ] Unit tests
- [ ] Integration tests
- [ ] GitHub Actions CI/CD
- [ ] ARCHITECTURE.md final review
