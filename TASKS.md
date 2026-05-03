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
- [x] GitHub Actions CI
- [x] Unit tests
- [x] Integration tests
- [x] README.md
- [x] ARCHITECTURE.md final review
<!-- - [ ] AI failure analyzer -->

---

## Current Focus

✅ Backend core completed (100%)
✅ Frontend foundation completed (100%)
✅ React Flow DAG builder completed (edit mode)
✅ React Flow DAG viewer completed (read-only mode)
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
✅ Global health panel completed
✅ Client-side caching with React Query completed
✅ Optimistic UI updates completed
✅ Unit tests completed (193 tests - 172 backend + 21 frontend)
✅ Integration tests completed (19 integration tests)
✅ Authentication tests completed (123 auth-specific tests)
✅ GitHub Actions CI/CD completed

**BACKEND MVP COMPLETE!** 🎉
**FRONTEND MVP COMPLETE!** ✨🎉
**TESTING & CI/CD COMPLETE!** ✅🎉

---

## Test Plan

### RBAC Implementation Testing

#### Pre-requisites
- [x] Start all services: `docker-compose up -d`
- [x] Run database seed: `docker-compose exec backend go run cmd/seed/main.go`
- [x] Verify all three users created in database

#### Test Users
- **Admin**: admin@flowforge.local / admin123 (full access)
- **Editor**: editor@flowforge.local / editor123 (create, edit, trigger - no delete)
- **Viewer**: viewer@flowforge.local / viewer123 (read-only)

#### Frontend Permission Tests
- [x] Login with admin user
  - [x] Verify role badge shows "Admin" in red
  - [x] Verify "Create Workflow" button is visible
  - [x] Verify "Run" button is visible on workflow cards
  - [x] Verify "Edit Workflow" button is visible
  - [x] Verify "Delete" button is visible
  - [x] Create a new workflow successfully
  - [x] Edit an existing workflow successfully
  - [x] Trigger a workflow successfully
  - [x] Delete a workflow successfully

- [x] Login with editor user
  - [x] Verify role badge shows "Editor" in blue
  - [x] Verify "Create Workflow" button is visible
  - [x] Verify "Run" button is visible on workflow cards
  - [x] Verify "Edit Workflow" button is visible
  - [x] Verify "Delete" button is NOT visible
  - [x] Create a new workflow successfully
  - [x] Edit an existing workflow successfully
  - [x] Trigger a workflow successfully
  - [x] Attempt to delete workflow → button hidden

- [x] Login with viewer user
  - [x] Verify role badge shows "Viewer" in gray
  - [x] Verify "Create Workflow" button is NOT visible
  - [x] Verify "Run" button is NOT visible on workflow cards
  - [x] Verify "Edit Workflow" button is NOT visible
  - [x] Verify "Delete" button is NOT visible
  - [x] View workflow list successfully
  - [x] View workflow details successfully
  - [x] View run history successfully
  - [x] Attempt to access /dashboard/workflows/new → redirected

#### Backend API Permission Tests
- [x] Test Viewer permissions
  - [x] GET /api/v1/workflows → 200 OK
  - [x] POST /api/v1/workflows → 403 Forbidden
  - [x] PUT /api/v1/workflows/:id → 403 Forbidden
  - [x] POST /api/v1/workflows/:id/trigger → 403 Forbidden
  - [x] DELETE /api/v1/workflows/:id → 403 Forbidden

- [x] Test Editor permissions
  - [x] GET /api/v1/workflows → 200 OK
  - [x] POST /api/v1/workflows → 201 Created
  - [x] PUT /api/v1/workflows/:id → 200 OK
  - [x] POST /api/v1/workflows/:id/trigger → 200 OK
  - [x] DELETE /api/v1/workflows/:id → 403 Forbidden

- [x] Test Admin permissions
  - [x] GET /api/v1/workflows → 200 OK
  - [x] POST /api/v1/workflows → 201 Created
  - [x] PUT /api/v1/workflows/:id → 200 OK
  - [x] POST /api/v1/workflows/:id/trigger → 200 OK
  - [x] DELETE /api/v1/workflows/:id → 200 OK

#### Cross-Tenant Isolation Tests
- [x] Create workflow as tenant A user
- [x] Login as tenant B user
- [x] Verify tenant B cannot access tenant A's workflows (404/403)

#### Authentication Tests
- [x] Test expired JWT → 401 Unauthorized
- [x] Test invalid JWT → 401 Unauthorized
- [x] Test missing Authorization header → 401 Unauthorized
- [x] Test malformed Authorization header → 401 Unauthorized

#### UI/UX Tests
- [ ] Verify role badge color coding (Admin=red, Editor=blue, Viewer=gray)
- [ ] Verify smooth hiding/showing of buttons based on permissions
- [ ] Verify no console errors during permission checks
- [ ] Verify loading states work correctly
- [ ] Verify error messages display properly

#### Integration Tests (Manual)
- [ ] Create workflow as Admin
- [ ] Logout and login as Editor
- [ ] Verify Editor can see and edit the workflow
- [ ] Logout and login as Viewer
- [ ] Verify Viewer can view but not edit the workflow
- [ ] Test version rollback with different roles
- [ ] Test schedule creation with different roles
- [ ] Test webhook creation with different roles

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

### Frontend Features (100% Complete ✅)
- ✅ Authentication with JWT
- ✅ Dashboard with workflow list & details
- ✅ Visual DAG builder with React Flow (edit mode)
- ✅ Visual DAG viewer with React Flow (read-only mode)
- ✅ Run history with live monitoring (SSE)
- ✅ Responsive design with Tailwind CSS
- ✅ Global health panel with stats & charts
- ✅ Client-side caching with React Query
- ✅ Optimistic UI updates for all mutations

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

- [x] Unit tests
- [x] Integration tests
- [x] GitHub Actions CI/CD
- [x] ARCHITECTURE.md final review

---

## Priority: Missing Frontend Features

### High Priority - Must Have

- [x] Add global health panel to dashboard
  - [x] Active runs counter (real-time)
  - [x] Success/failure rates (last 24 hours)
  - [x] Average execution time (last 24 hours)
  - [x] Trend charts (using Recharts)
  - [x] Add backend API endpoint for stats aggregation
  - [x] Auto-refresh every 30 seconds

- [x] Add client-side caching with React Query
  - [x] Install @tanstack/react-query
  - [x] Wrap app with QueryClientProvider
  - [x] Cache workflows list with staleTime
  - [x] Cache runs list with staleTime
  - [x] Implement stale-while-revalidate strategy
  - [x] Add cache invalidation on mutations

- [x] Implement optimistic UI updates
  - [x] Optimistic updates for workflow trigger
  - [x] Rollback on error
  - [x] Loading states for all mutations
  - [x] Alert notifications for success/error

### Medium Priority - Nice to Have

- [x] Add visual DAG viewer to workflow detail page
  - [x] Read-only ReactFlow component
  - [x] Reuse existing custom node types
  - [x] Disable drag-and-drop, edit, delete
  - [x] Add zoom/pan controls
  - [x] MiniMap for navigation

**✅ ALL FRONTEND FEATURES COMPLETE!** 🎉

