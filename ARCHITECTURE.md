# FlowForge Architecture

## Overview

FlowForge is a real-time multi-tenant workflow orchestration platform inspired by Zapier + GitHub Actions. It enables users to create, execute, and monitor complex workflows with a visual DAG (Directed Acyclic Graph) builder.

**Tech Stack:**
- **Backend:** Golang + Fiber web framework
- **Database:** PostgreSQL with migrations
- **Cache/Queue:** Redis
- **Frontend:** Next.js 16 + React 19 + Tailwind CSS 4
- **Realtime:** Server-Sent Events (SSE)
- **Visualization:** React Flow for DAG builder/viewer
- **Testing:** Go testing + testify, Jest + React Testing Library
- **CI/CD:** GitHub Actions
- **Deployment:** Docker + Docker Compose

---

## Project Structure

```
flowforge/
├── backend/           # Go backend service
├── frontend/          # Next.js frontend app
├── .github/           # GitHub Actions workflows
├── ARCHITECTURE.md    # This file
├── TASKS.md           # Development tasks & progress
└── README.md          # Project documentation
```

---

## Backend Architecture

### Directory Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go           # Application entry point
├── internal/
│   ├── auth/                 # JWT authentication & management
│   │   ├── jwt.go            # JWT token generation & validation
│   │   └── jwt_test.go       # JWT tests
│   ├── config/               # Configuration management
│   │   └── config.go         # Env-based config loading
│   ├── dag/                  # DAG validation & processing
│   │   ├── validator.go      # Cycle detection & topological sort
│   │   └── validator_test.go # DAG tests
│   ├── db/                   # Database connections
│   │   ├── postgres.go       # PostgreSQL connection pool
│   │   └── redis.go          # Redis client
│   ├── execution/            # Workflow execution engine
│   │   ├── executor.go       # Execute workflow steps
│   │   └── worker.go         # Background worker
│   ├── handlers/             # HTTP request handlers
│   │   ├── auth.go           # Auth endpoints
│   │   ├── health.go         # Health check
│   │   ├── runs.go           # Run endpoints
│   │   ├── schedule.go       # Schedule endpoints
│   │   ├── stats.go          # Statistics endpoints
│   │   └── workflow.go       # Workflow endpoints
│   ├── middleware/           # Fiber middleware
│   │   ├── auth.go           # JWT authentication
│   │   ├── role.go           # Role-based access control
│   │   └── ratelimit.go      # Redis rate limiting
│   ├── migrate/              # Database migrations
│   │   └── migrate.go        # Migration runner
│   ├── models/               # Data models
│   │   ├── schedule.go       # Schedule model
│   │   ├── user.go           # User model
│   │   ├── webhook.go        # Webhook model
│   │   ├── workflow.go       # Workflow models
│   │   └── workflow_run.go   # Run & step models
│   ├── repository/           # Database access layer
│   │   ├── schedule.go       # Schedule repository
│   │   ├── user.go           # User repository
│   │   ├── webhook.go        # Webhook repository
│   │   ├── workflow.go       # Workflow repository
│   │   └── workflow_run.go   # Run repository
│   ├── scheduler/            # Cron scheduler
│   │   └── scheduler.go      # Background cron jobs
│   ├── server/               # HTTP server setup
│   │   └── server.go         # Fiber app & routing
│   └── validator/            # Input validation
│       ├── validator.go      # Validation functions
│       └── validator_test.go # Validation tests
├── migrations/               # SQL migrations
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_add_schedules_and_webhooks.up.sql
│   └── 000002_add_schedules_and_webhooks.down.sql
├── .env.example              # Environment variables template
├── Dockerfile                # Multi-stage Docker build
├── Makefile                  # Development commands
├── go.mod                    # Go dependencies
└── go.sum                    # Go dependency lock
```

### Core Components

#### 1. **Server** (`internal/server/`)
- Fiber web application setup
- Route registration & middleware
- HTTP server configuration
- Timeout & idle settings

#### 2. **Handlers** (`internal/handlers/`)
- **AuthHandler:** Login, user info
- **WorkflowHandler:** CRUD, trigger, version rollback
- **RunHandler:** List, get details, SSE stream
- **ScheduleHandler:** CRUD for cron schedules
- **WebhookHandler:** CRUD & public trigger endpoint
- **StatsHandler:** Health statistics & metrics

#### 3. **Repository Layer** (`internal/repository/`)
- Abstracts database operations
- Handles tenant isolation (automatic tenant_id filtering)
- Provides transaction support
- Pagination & filtering helpers

#### 4. **Middleware** (`internal/middleware/`)
- **Auth:** JWT validation & user context
- **Role:** RBAC enforcement (admin, editor, viewer)
- **RateLimit:** Redis-based sliding window rate limiting

#### 5. **Execution Engine** (`internal/execution/`)
- **Executor:** Execute workflow steps in topological order
- **Worker:** Background worker for async execution
- **Retry Logic:** Exponential backoff for failed steps
- **Timeout Enforcement:** Per-step and per-workflow timeouts

#### 6. **DAG Validator** (`internal/dag/`)
- **Cycle Detection:** Prevent infinite loops
- **Topological Sort:** Determine execution order
- **Execution Levels:** Calculate parallel execution groups

#### 7. **Scheduler** (`internal/scheduler/`)
- Background cron job manager
- Checks due schedules every minute
- Triggers workflows automatically
- Updates next_run_at timestamps

#### 8. **Validator** (`internal/validator/`)
- Email format validation
- UUID validation
- Cron expression validation
- SQL injection prevention
- Field length & format checks

---

## Frontend Architecture

### Directory Structure

```
frontend/
├── src/
│   ├── app/                      # Next.js App Router
│   │   ├── dashboard/            # Dashboard page
│   │   │   └── page.tsx          # Main dashboard with stats
│   │   ├── login/                # Login page
│   │   │   └── page.tsx          # Login form
│   │   ├── workflows/            # Workflow pages
│   │   │   ├── [id]/             # Workflow detail
│   │   │   │   └── page.tsx      # DAG viewer + editor
│   │   │   └── page.tsx          # Workflow list
│   │   ├── runs/                 # Run pages
│   │   │   └── [id]/             # Run detail
│   │   │       └── page.tsx      # Run viewer with SSE
│   │   ├── layout.tsx            # Root layout
│   │   └── page.tsx              # Home/redirect
│   ├── components/               # React components
│   │   ├── nodes/                # React Flow custom nodes
│   │   │   ├── DelayNode.tsx
│   │   │   ├── HTTPNode.tsx
│   │   │   └── ScriptNode.tsx
│   │   ├── DashboardLayout.tsx   # Dashboard layout wrapper
│   │   └── WorkflowList.tsx      # Workflow list component
│   ├── lib/                      # Utilities & libraries
│   │   ├── __tests__/            # Unit tests
│   │   │   ├── api.test.ts       # API client tests
│   │   │   └── auth.test.tsx     # Auth context tests
│   │   ├── api.ts                # API client
│   │   └── auth.tsx              # Auth context & hooks
│   └── ...
├── jest.config.js                # Jest configuration
├── jest.setup.js                 # Test setup
├── next.config.ts                # Next.js config
├── tailwind.config.ts            # Tailwind CSS config
├── package.json                  # Dependencies
├── Dockerfile                    # Multi-stage Docker build
└── tsconfig.json                 # TypeScript config
```

### Core Components

#### 1. **API Client** (`src/lib/api.ts`)
- HTTP client with JWT authentication
- Automatic token management
- Error handling
- Typed request/response interfaces

#### 2. **Auth Context** (`src/lib/auth.tsx`)
- JWT token persistence (localStorage)
- User authentication state
- Login/logout functions
- Protected route wrapper

#### 3. **DAG Builder/Viewer** (React Flow)
- **Edit Mode:** Create and modify workflows
  - Drag-and-drop nodes
  - Connect nodes with edges
  - Configure node properties
- **View Mode:** Read-only visualization
  - MiniMap for navigation
  - Zoom and pan controls
  - Execution status overlays

#### 4. **Real-time Monitoring** (SSE)
- Live run updates via Server-Sent Events
- Step-by-step execution display
- Status badges (running, success, failed)
- Auto-scrolling log viewer

#### 5. **Global Health Panel**
- Active runs counter
- Success/failure rates (24h)
- Average execution time
- Trend charts (Recharts)
- Auto-refresh every 30s

#### 6. **Client-Side Caching** (React Query)
- Cache workflows & runs
- Stale-while-revalidate strategy
- Optimistic updates
- Cache invalidation on mutations

---

## Database Schema

### Core Tables

```sql
-- Tenants (multi-tenant isolation)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Users (authentication & roles)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL, -- 'admin', 'editor', 'viewer'
    created_at TIMESTAMP DEFAULT NOW()
);

-- Workflows (DAG definitions)
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    definition JSONB NOT NULL, -- {nodes: [], edges: []}
    timeout_seconds INT DEFAULT 300,
    active BOOLEAN DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Workflow Versions (version history)
CREATE TABLE workflow_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id),
    version INT NOT NULL,
    definition JSONB NOT NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(workflow_id, version)
);

-- Workflow Runs (execution records)
CREATE TABLE workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    status VARCHAR(50) NOT NULL, -- 'pending', 'running', 'success', 'failed', 'cancelled'
    error TEXT,
    triggered_by VARCHAR(100) NOT NULL, -- 'manual', 'webhook', 'scheduler'
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Run Steps (individual step execution)
CREATE TABLE workflow_run_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES workflow_runs(id),
    step_id VARCHAR(255) NOT NULL, -- Node ID from definition
    status VARCHAR(50) NOT NULL, -- 'pending', 'running', 'success', 'failed', 'skipped'
    input JSONB,
    output JSONB,
    error TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Schedules (cron triggers)
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    cron_expression VARCHAR(100) NOT NULL,
    active BOOLEAN DEFAULT true,
    next_run_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Webhooks (HTTP triggers)
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    path VARCHAR(255) UNIQUE NOT NULL, -- UUID for URL path
    secret VARCHAR(255) NOT NULL, -- HMAC secret
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Execution Logs (high-volume, separate store)
CREATE TABLE workflow_logs (
    id BIGSERIAL PRIMARY KEY,
    run_id UUID NOT NULL REFERENCES workflow_runs(id),
    level VARCHAR(20) NOT NULL, -- 'info', 'warning', 'error', 'debug'
    message TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_workflows_tenant ON workflows(tenant_id);
CREATE INDEX idx_workflows_active ON workflows(active) WHERE active = true;
CREATE INDEX idx_runs_tenant ON workflow_runs(tenant_id);
CREATE INDEX idx_runs_workflow ON workflow_runs(workflow_id);
CREATE INDEX idx_runs_status ON workflow_runs(status);
CREATE INDEX idx_runs_created ON workflow_runs(created_at DESC);
CREATE INDEX idx_steps_run ON workflow_run_steps(run_id);
CREATE INDEX idx_steps_status ON workflow_run_steps(status);
CREATE INDEX idx_logs_run ON workflow_logs(run_id);
CREATE INDEX idx_logs_created ON workflow_logs(created_at DESC);
CREATE INDEX idx_schedules_tenant ON schedules(tenant_id);
CREATE INDEX idx_schedules_active ON schedules(active) WHERE active = true;
CREATE INDEX idx_schedules_next_run ON schedules(next_run_at) WHERE active = true;
CREATE INDEX idx_webhooks_tenant ON webhooks(tenant_id);
CREATE INDEX idx_webhooks_active ON webhooks(active) WHERE active = true;
CREATE INDEX idx_webhooks_path ON webhooks(path);
```

### Tenant Isolation

All tenant-owned tables include:
- `tenant_id` column (foreign key to tenants)
- Index on `tenant_id` for query performance
- Repository layer automatically filters by tenant_id
- Middleware validates tenant access

---

## API Routes

### Public Routes

```
GET  /health                          # Health check
POST /api/v1/auth/login              # Authenticate user
POST /webhooks/:path                 # Public webhook trigger (with signature)
```

### Protected Routes (JWT Required)

#### Authentication
```
GET  /api/v1/auth/me                 # Get current user info
```

#### Workflows
```
GET    /api/v1/workflows             # List workflows (paginated, filtered)
POST   /api/v1/workflows             # Create workflow (editor/admin)
GET    /api/v1/workflows/:id         # Get workflow details
PUT    /api/v1/workflows/:id         # Update workflow (editor/admin)
DELETE /api/v1/workflows/:id         # Delete workflow (admin only)
POST   /api/v1/workflows/:id/trigger # Trigger workflow (editor/admin)
GET    /api/v1/workflows/:id/versions # List version history
POST   /api/v1/workflows/:id/rollback/:version # Restore version (editor/admin)
```

#### Runs
```
GET /api/v1/runs                     # List runs (paginated, filtered)
GET /api/v1/runs/:id                 # Get run details
GET /api/v1/runs/:id/stream          # SSE stream for live updates
```

#### Schedules
```
GET    /api/v1/schedules             # List schedules
POST   /api/v1/workflows/:id/schedule # Create schedule (editor/admin)
DELETE /api/v1/schedules/:id         # Delete schedule (editor/admin)
```

#### Webhooks
```
GET    /api/v1/webhooks              # List webhooks
POST   /api/v1/workflows/:id/webhook # Create webhook (editor/admin)
DELETE /api/v1/webhooks/:id          # Delete webhook (editor/admin)
```

#### Statistics
```
GET /api/v1/stats/health             # Health statistics (24h metrics)
```

### Rate Limiting

- **Auth endpoints:** 10 req/min (by IP)
- **Read endpoints:** 100 req/min (by user)
- **Write endpoints:** 30 req/min (by user)
- **Trigger endpoints:** 10 req/min (by user)

Headers returned:
- `X-RateLimit-Limit`: Request limit
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Reset time

---

## Authentication & Authorization

### JWT Flow

1. **Login:**
   ```
   POST /api/v1/auth/login
   Body: { email, password }
   Response: { token, user }
   ```

2. **Token Storage:**
   - Client stores JWT in localStorage
   - Subsequent requests include `Authorization: Bearer <token>`

3. **Token Validation:**
   - Auth middleware verifies JWT signature
   - Extracts user_id, tenant_id, role
   - Sets context for request handlers

### Role-Based Access Control (RBAC)

| Role | Permissions |
|------|-------------|
| **viewer** | View workflows, runs, stats (read-only) |
| **editor** | viewer + Create/edit/trigger workflows, manage schedules & webhooks |
| **admin** | editor + Delete workflows, manage users |

### Middleware Chain

```
Request → Rate Limit → Auth (JWT) → Role Check → Handler
```

---

## Workflow Definition

### JSON Schema

```json
{
  "nodes": [
    {
      "id": "node-1",
      "type": "http|delay|script|condition",
      "name": "Fetch Data",
      "config": {
        // Type-specific config
      },
      "position": {
        "x": 100,
        "y": 100
      }
    }
  ],
  "edges": [
    {
      "id": "edge-1",
      "source": "node-1",
      "target": "node-2"
    }
  ]
}
```

### Node Types

#### 1. **HTTP Node**
```json
{
  "type": "http",
  "config": {
    "url": "https://api.example.com/data",
    "method": "GET|POST|PUT|DELETE",
    "headers": {
      "Authorization": "Bearer <token>"
    },
    "body": {}
  }
}
```

#### 2. **Delay Node**
```json
{
  "type": "delay",
  "config": {
    "seconds": 5
  }
}
```

#### 3. **Script Node**
```json
{
  "type": "script",
  "config": {
    "script": "return input.data * 2"
  }
}
```

#### 4. **Condition Node**
```json
{
  "type": "condition",
  "config": {
    "expression": "input.status == 200"
  }
}
```

### Execution Rules

1. **No Cycles:** DAG must be acyclic (validated before save)
2. **Topological Order:** Execute in dependency order
3. **Parallel Execution:** Independent steps run concurrently
4. **Retry:** Failed steps retry with exponential backoff (max 3 retries)
5. **Timeout:** Per-step timeout (default 30s), per-workflow timeout (default 300s)
6. **Fail Fast:** Workflow stops on first failure (configurable)

---

## Execution Engine

### Workflow Lifecycle

```
Trigger → Validate → Queue → Execute → Update Status → Notify
```

### Step Execution

```
For each step in topological order:
  1. Validate step config
  2. Execute step (HTTP call, delay, script, etc.)
  3. Retry on failure (exponential backoff)
  4. Record result (success/failed)
  5. Pass output to next steps
```

### Status Flow

```
pending → running → success
                └→ failed
                └→ cancelled
```

### Real-time Updates (SSE)

Events emitted during execution:
- `run_started`: Workflow execution started
- `step_started`: Step execution started
- `step_success`: Step completed successfully
- `step_failed`: Step failed (after retries)
- `run_completed`: Workflow finished (success or failed)

Client connects to:
```
GET /api/v1/runs/:id/stream
Accept: text/event-stream
```

---

## Scheduler (Cron)

### How It Works

1. Background worker checks every minute for due schedules
2. Finds schedules where `next_run_at <= NOW()` and `active = true`
3. Triggers associated workflow
4. Calculates next run time using cron expression
5. Updates `next_run_at` timestamp

### Cron Format

Standard 5-field cron expression:
```
* * * * *
│ │ │ │ │
│ │ │ │ └─── Day of week (0-6, Sunday = 0)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
```

Examples:
- `0 * * * *` - Every hour
- `*/30 * * * *` - Every 30 minutes
- `0 9 * * 1-5` - 9 AM on weekdays
- `0 0 * * 0` - Midnight on Sunday

---

## Webhooks

### How It Works

1. User creates webhook for workflow
2. System generates unique path (UUID)
3. System generates HMAC secret
4. External service sends POST to `/webhooks/:path`
5. Server validates signature (HMAC SHA256)
6. Triggers associated workflow

### Signature Verification

```
Signature = HMAC-SHA256(secret, request_body)
Header: X-Webhook-Signature: <signature>
```

### Usage

Create webhook:
```bash
POST /api/v1/workflows/:id/webhook
Response: { path: "abc-123-def", secret: "xyz-789-uvw" }
```

Trigger webhook:
```bash
POST /webhooks/abc-123-def
Header: X-Webhook-Signature: <calculated_hmac>
Body: { data: "any json data" }
```

---

## Rate Limiting

### Strategy

Redis-based sliding window rate limiter:
- Tracks request timestamps per user/IP
- Counts requests in time window
- Rejects when limit exceeded

### Configuration

```go
rateLimit.AddConfig(
    name: "auth",
    limit: 10,
    window: 1 minute,
    keyStrategy: ByIP
)
```

### Response (Rate Limited)

```json
Status: 429 Too Many Requests
{
  "error": "Rate limit exceeded",
  "retry_after": 45
}
```

Headers:
- `X-RateLimit-Limit`: 10
- `X-RateLimit-Remaining`: 0
- `X-RateLimit-Reset`: Unix timestamp

---

## Input Validation

### Validation Rules

| Field | Rules |
|-------|-------|
| **Email** | Valid email format, RFC 5322 |
| **UUID** | v4 format, 8-4-4-4-12 hex digits |
| **Cron** | 5-field cron expression |
| **Workflow Name** | 1-255 chars, alphanumeric + spaces + _- |
| **Workflow Description** | Max 5000 chars |
| **SQL Injection** | Reject dangerous patterns (`; DROP`, `UNION`, etc.) |

### Sanitization

- Trim whitespace from all inputs
- Escape HTML entities to prevent XSS
- Validate against allowlists (node types, HTTP methods)
- Reject invalid UTF-8 sequences

---

## Testing

### Backend Tests (Go)

**Framework:** Go testing + testify

**Coverage:**
- DAG validator: 34 tests
- Input validator: 40 tests
- JWT auth: 13 tests
- **Total: 87 tests**

**Run tests:**
```bash
cd backend
go test ./... -v
go test ./... -race -cover
```

**Test structure:**
```go
func TestValidator_DetectCycles_SimpleCycle(t *testing.T) {
    v := NewValidator()
    // Arrange
    def := models.WorkflowDef{...}
    // Act
    err := v.Validate(def)
    // Assert
    assert.Error(t, err)
    assert.Equal(t, ErrCycleDetected, err)
}
```

### Frontend Tests (Jest)

**Framework:** Jest + React Testing Library

**Coverage:**
- API client: 14 tests
- Auth context: 7 tests
- **Total: 21 tests**

**Run tests:**
```bash
cd frontend
npm test
npm test:watch
npm test:coverage
```

**Test structure:**
```typescript
describe('API', () => {
  it('should login successfully', async () => {
    const result = await apiClient.login({ email, password });
    expect(result.token).toBe('test-token');
  });
});
```

### CI/CD (GitHub Actions)

**Backend Workflow:**
- Triggered on backend changes
- Runs on Ubuntu + Go 1.26
- Executes `go vet` and `go test`
- Uploads coverage to Codecov
- Enforces 50% coverage threshold

**Frontend Workflow:**
- Triggered on frontend changes
- Runs on Ubuntu + Node.js 20
- Executes `npm test` with coverage
- Uploads coverage to Codecov
- Enforces 50% coverage threshold

---

## Deployment

### Local Development (Docker Compose)

```bash
# Start all services
docker-compose up

# Backend: http://localhost:3000
# Frontend: http://localhost:3001
# PostgreSQL: localhost:5432
# Redis: localhost:6379
```

### Production Deployment

**Backend:**
```dockerfile
# Multi-stage build
FROM golang:1.26 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o flowforge cmd/api/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/flowforge /usr/local/bin/
EXPOSE 3000
CMD ["flowforge"]
```

**Frontend:**
```dockerfile
# Multi-stage build
FROM node:20 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json
EXPOSE 3000
CMD ["npm", "start"]
```

### Environment Variables

**Backend (.env):**
```env
# Server
PORT=3000
JWT_SECRET=your-secret-key

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=flowforge
DB_PASSWORD=flowforge
DB_NAME=flowforge

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
```

**Frontend (.env.local):**
```env
NEXT_PUBLIC_API_URL=http://localhost:3000/api/v1
```

---

## Frontend Pages

### Page Routes

| Route | Component | Description |
|-------|-----------|-------------|
| `/` | `app/page.tsx` | Home → redirect to dashboard |
| `/login` | `app/login/page.tsx` | Login form |
| `/dashboard` | `app/dashboard/page.tsx` | Dashboard with health panel |
| `/workflows` | `app/workflows/page.tsx` | Workflow list |
| `/workflows/[id]` | `app/workflows/[id]/page.tsx` | Workflow detail + DAG editor/viewer |
| `/runs/[id]` | `app/runs/[id]/page.tsx` | Run details + SSE live updates |

### Layout

- **DashboardLayout:** Wrapper for authenticated pages
  - Sidebar navigation
  - Header with user menu
  - Global health panel (stats + charts)

---

## Performance Optimizations

### Database

1. **Indexes:** All foreign keys, frequently queried columns
2. **Connection Pooling:** PostgreSQL pool (max 25 connections)
3. **Partitioning:** Consider partitioning `workflow_logs` by date for scale

### Caching

1. **Redis:** Rate limiting, session data
2. **React Query:** Client-side caching for workflows & runs
3. **Stale-While-Revalidate:** Show cached data, refresh in background

### Execution

1. **Parallel Steps:** Execute independent steps concurrently
2. **Background Worker:** Async execution with Redis queue
3. **Timeouts:** Prevent hanging requests

---

## Security Considerations

### Authentication

- JWT with HS256 signing
- Token expiration (24 hours)
- Secure password hashing (bcrypt)
- Password requirements (min 8 chars)

### Authorization

- Tenant isolation enforced at database level
- Role-based access control (RBAC)
- Middleware checks on all protected routes

### Input Validation

- Whitelist validation (node types, HTTP methods)
- SQL injection prevention (sanitization)
- XSS prevention (HTML escaping)
- CSRF protection (same-origin checks)

### Rate Limiting

- Prevents brute force attacks
- Per-IP limits for auth endpoints
- Per-user limits for API endpoints
- Redis-based for distributed systems

### Webhook Security

- HMAC SHA256 signature verification
- Secret per webhook
- UUID paths prevent guessing

---

## Future Enhancements

### Planned Features

1. **Integration Tests:** End-to-end API and frontend tests
2. **AI Failure Analyzer:** Automatic error analysis & suggestions
3. **More Node Types:** Email, database, file storage, etc.
4. **Workflow Templates:** Pre-built workflow patterns
5. **Version Diff:** Visual diff between workflow versions
6. **Audit Log:** Track all workflow modifications
7. **Webhook Retry:** Retry failed webhook deliveries
8. **Conditional Routing:** Complex branch logic
9. **Sub-workflows:** Reusable workflow components
10. **Metrics Export:** Prometheus, Grafana dashboards

### Scalability Considerations

1. **Horizontal Scaling:** Multiple backend instances with shared Redis/PostgreSQL
2. **Worker Pool:** Dedicated execution workers
3. **Message Queue:** Replace Redis with RabbitMQ/Kafka for scale
4. **Database Sharding:** Partition by tenant_id for multi-tenant scale
5. **CDN:** Serve static assets via CDN
6. **Frontend Caching:** Service worker for offline support

---

## Development Workflow

### Backend Development

```bash
# Run backend
cd backend
make dev     # Hot reload with air
make test    # Run tests
make build   # Build binary

# Run migrations
make migrate_up
make migrate_down
```

### Frontend Development

```bash
# Run frontend
cd frontend
npm run dev       # Hot reload
npm run build     # Production build
npm run start     # Production server
npm run test      # Run tests
npm run lint      # ESLint
```

### Git Workflow

1. Create feature branch: `git checkout -b feature/xxx`
2. Make changes and commit
3. Push to remote: `git push -u origin feature/xxx`
4. Create pull request
5. CI runs automatically (tests, coverage)
6. Merge to main after approval

---

## Architecture Principles

### 1. **Simplicity First**
- Clear, readable code over clever tricks
- Straightforward data flow
- Minimal dependencies

### 2. **Multi-Tenant by Default**
- Every resource belongs to a tenant
- Tenant isolation enforced at all layers
- No cross-tenant data leakage

### 3. **API-First Design**
- RESTful API conventions
- Consistent response formats
- Versioned API (`/api/v1/`)

### 4. **Realtime Feedback**
- SSE for live updates
- Optimistic UI updates
- Progress indicators

### 5. **Fail Gracefully**
- Validation before execution
- Clear error messages
- Retry logic for transient failures
- Timeouts prevent hanging

### 6. **Test Coverage**
- Unit tests for critical logic
- Integration tests for API contracts
- CI/CD automation

---

## Summary

FlowForge is a complete multi-tenant workflow automation platform with:

✅ **Backend:** Go + Fiber, PostgreSQL, Redis
✅ **Frontend:** Next.js 16, React 19, Tailwind CSS 4
✅ **Core Features:** DAG builder, execution engine, SSE monitoring
✅ **Advanced Features:** Scheduling, webhooks, versioning, RBAC
✅ **Testing:** 108 unit tests, CI/CD with GitHub Actions
✅ **Deployment:** Docker, Docker Compose ready
✅ **Security:** JWT auth, tenant isolation, rate limiting, input validation

**MVP COMPLETE!** 🎉
