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

✅ Backend core completed (100%)
⚠️ Frontend foundation completed (70%)
✅ React Flow DAG builder completed (edit mode)
✅ SSE live monitoring completed
✅ Docker deployment completed
✅ Documentation completed
✅ Pagination & filtering completed
✅ Input validation & sanitization completed
✅ Rate limiting completed
✅ RBAC enforcement completed
✅ Workflow version rollback completed
✅ Scheduled/cron triggering completed
✅ Webhook triggering completed

**BACKEND MVP COMPLETE!** 🎉
**Frontend enhancements needed for full parity** 🚧

---

## Summary

### Backend Features (100% Complete ✅)
- ✅ Full CRUD workflows with versioning & rollback
- ✅ Workflow triggering (manual, scheduled, webhook)
- ✅ Pagination & filtering on all list endpoints
- ✅ Multi-tenant isolation (strict separation)
- ✅ JWT authentication with role-based access control
- ✅ Comprehensive input validation & sanitization
- ✅ Rate limiting (Redis-based, configurable per endpoint)

### Frontend Features (70% Complete ⚠️)
- ✅ Authentication with JWT
- ✅ Dashboard with workflow list & details
- ✅ Visual DAG builder with React Flow (edit mode)
- ✅ Run history with live monitoring (SSE)
- ✅ Responsive design with Tailwind CSS
- ⚠️ Visual DAG viewer in workflow detail (read-only mode)
- ❌ Global health panel with stats
- ❌ Client-side caching
- ❌ Optimistic UI updates

### Infrastructure (100% Complete ✅)
- ✅ Docker multi-stage builds
- ✅ Docker Compose for local development
- ✅ PostgreSQL + Redis
- ✅ Database migrations

---

## Priority: Missing Core Backend Features

### High Priority - Must Have

- [x] Add workflow version rollback endpoint
  - [x] GET /api/v1/workflows/:id/versions (list all versions)
  - [x] POST /api/v1/workflows/:id/rollback/:version (restore specific version)
  - [x] Update repository with GetVersions and Rollback methods

- [x] Add scheduled/cron triggering
  - [x] Create schedules table (id, workflow_id, tenant_id, cron_expression, active, next_run_at)
  - [x] Add cron parser library (github.com/robfig/cron/v3)
  - [x] Create scheduler service that checks and triggers scheduled workflows
  - [x] Add API endpoints (POST /api/v1/workflows/:id/schedule, GET /api/v1/schedules, DELETE /api/v1/schedules/:id)

- [x] Add webhook triggering
  - [x] Create webhooks table (id, workflow_id, tenant_id, path, secret, active)
  - [x] Generate unique webhook URLs (e.g., /webhooks/{uuid})
  - [x] Add webhook handler with signature verification
  - [x] Add API endpoints (POST /api/v1/workflows/:id/webhook, GET /api/v1/webhooks, DELETE /api/v1/webhooks/:id)

- [x] Add pagination to all list endpoints
  - [x] Update ListWorkflows with limit/offset/cursor pagination
  - [x] Update ListRuns with limit/offset/cursor pagination
  - [x] Add response metadata (total, page, per_page, total_pages)

- [x] Add filtering to all list endpoints
  - [x] Workflows: filter by active, created_by, date range
  - [x] Runs: filter by status, workflow_id, triggered_by, date range
  - [x] Add query parameter parsing and validation

- [x] Add rate limiting middleware
  - [x] Redis-based rate limiter (sliding window)
  - [x] Configurable limits per endpoint/role
  - [x] Add rate limit headers to responses

- [x] Enforce role-based access control
  - [x] Apply Role middleware to all endpoints
  - [x] Viewer: read-only access
  - [x] Editor: create, update, trigger workflows
  - [x] Admin: full access including delete

- [x] Add comprehensive input validation
  - [x] Add field length limits (name: 255, description: 5000)
  - [x] Add format validation (email, UUID, cron expression)
  - [x] Add input sanitization (trim spaces, escape HTML)
  - [x] Add custom validator middleware

### Medium Priority

- [ ] Unit tests
- [ ] Integration tests
- [ ] GitHub Actions CI/CD
- [ ] ARCHITECTURE.md final review

---

## Priority: Missing Frontend Features

### High Priority - Must Have

- [ ] Add global health panel to dashboard
  - [ ] Active runs counter (real-time)
  - [ ] Success/failure rates (last 24 hours)
  - [ ] Average execution time (last 24 hours)
  - [ ] Trend charts (using Recharts or Chart.js)
  - [ ] Add backend API endpoint for stats aggregation
  - [ ] Poll or SSE for real-time updates

- [ ] Add client-side caching with React Query
  - [ ] Install @tanstack/react-query
  - [ ] Wrap app with QueryClientProvider
  - [ ] Cache workflows list with staleTime
  - [ ] Cache runs list with staleTime
  - [ ] Implement stale-while-revalidate strategy
  - [ ] Add cache invalidation on mutations

- [ ] Implement optimistic UI updates
  - [ ] Optimistic updates for workflow trigger
  - [ ] Rollback on error
  - [ ] Loading states for all mutations
  - [ ] Toast notifications for success/error

### Medium Priority - Nice to Have

- [ ] Add visual DAG viewer to workflow detail page
  - [ ] Read-only ReactFlow component
  - [ ] Reuse existing custom node types
  - [ ] Disable drag-and-drop, edit, delete
  - [ ] Add zoom/pan controls
  - [ ] MiniMap for navigation

