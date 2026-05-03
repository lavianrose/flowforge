# FlowForge

Real-time multi-tenant workflow orchestration platform inspired by Zapier + GitHub Actions.

## Features

- **Visual Workflow Builder**: Drag-and-drop DAG editor using React Flow
- **Real-time Execution**: Live monitoring with Server-Sent Events (SSE)
- **Multi-tenant**: Complete tenant isolation at database and API level
- **Role-based Access Control**: Admin, Editor, Viewer roles with middleware enforcement
- **Workflow Types**: HTTP requests, delays, scripts, conditions
- **Execution Engine**: Parallel processing with topological sorting
- **Version Control**: Automatic workflow versioning with rollback
- **Scheduling**: Cron-based workflow triggers
- **Webhooks**: HTTP-based triggers with HMAC signature verification
- **Rate Limiting**: Redis-based sliding window rate limiter
- **Health Dashboard**: 24-hour statistics with trend charts
- **Graceful Shutdown**: Proper server shutdown with signal handling and resource cleanup
- **Caching**: React Query with stale-while-revalidate strategy

## Tech Stack

### Backend

- **Language**: Go 1.26+
- **Framework**: Fiber (high-performance HTTP)
- **Database**: PostgreSQL 16
- **Cache/Queue**: Redis 7
- **Realtime**: Server-Sent Events (SSE)

### Frontend

- **Framework**: Next.js 16 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS 4
- **DAG Editor**: React Flow
- **State**: React Context API + React Query
- **Charts**: Recharts

### Infrastructure

- **Containerization**: Docker multi-stage builds
- **Orchestration**: Docker Compose
- **CI/CD**: GitHub Actions
- **Testing**: 193+ tests with 50% coverage threshold

## Quick Start

### Prerequisites

- Docker & Docker Compose

### Using Docker

```bash
# Clone the repository
git clone https://github.com/lavianrose/flowforge.git
cd flowforge

# Start all services
docker-compose up -d

# Access the application
# Frontend: http://localhost:3001
# Backend API: http://localhost:3000
```

## Default Credentials

| Role   | Email                  | Password  | Access                              |
| ------ | ---------------------- | --------- | ----------------------------------- |
| Admin  | admin@flowforge.local  | admin123  | Full access (CRUD, trigger, delete) |
| Editor | editor@flowforge.local | editor123 | Create, edit, trigger workflows     |
| Viewer | viewer@flowforge.local | viewer123 | Read-only access                    |

## API Documentation

### Authentication

#### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@flowforge.local",
  "password": "admin123"
}
```

#### Get Current User

```http
GET /api/v1/auth/me
Authorization: Bearer <token>
```

### Workflows

#### List Workflows

```http
GET /api/v1/workflows?page=1&limit=20
Authorization: Bearer <token>
```

#### Create Workflow

```http
POST /api/v1/workflows
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "My Workflow",
  "description": "Does something cool",
  "definition": {
    "nodes": [
      {
        "id": "node1",
        "type": "http",
        "name": "HTTP Request",
        "config": {
          "url": "https://api.example.com",
          "method": "GET"
        },
        "position": { "x": 100, "y": 100 }
      }
    ],
    "edges": []
  },
  "timeout_seconds": 300
}
```

#### Trigger Workflow

```http
POST /api/v1/workflows/{id}/trigger
Authorization: Bearer <token>
```

#### Get Workflow Versions

```http
GET /api/v1/workflows/{id}/versions
Authorization: Bearer <token>
```

#### Rollback Workflow

```http
POST /api/v1/workflows/{id}/rollback/{version}
Authorization: Bearer <token>
```

### Runs

#### List Runs

```http
GET /api/v1/runs?page=1&limit=20&status=completed
Authorization: Bearer <token>
```

#### Get Run Details

```http
GET /api/v1/runs/{id}
Authorization: Bearer <token>
```

#### Stream Run (SSE)

```http
GET /api/v1/runs/{id}/stream
Authorization: Bearer <token>
Accept: text/event-stream
```

### Schedules

#### Create Schedule

```http
POST /api/v1/workflows/{id}/schedule
Authorization: Bearer <token>
Content-Type: application/json

{
  "cron": "*/5 * * * *",
  "enabled": true
}
```

#### List Schedules

```http
GET /api/v1/schedules
Authorization: Bearer <token>
```

### Webhooks

#### Create Webhook

```http
POST /api/v1/workflows/{id}/webhook
Authorization: Bearer <token>
```

#### Trigger via Webhook

```http
POST /webhooks/{path}
X-Hub-Signature-256: sha256=<hmac_signature>
Content-Type: application/json

{
  "payload": "data"
}
```

### Statistics

#### Health Statistics

```http
GET /api/v1/stats/health
Authorization: Bearer <token>
```

## Development

### Backend Commands

```bash
make build        # Build binaries
make run          # Start server
make migrate-up   # Run migrations
make migrate-down # Rollback migrations
make seed         # Seed users
make clean        # Clean build artifacts
```

### Frontend Commands

```bash
npm run dev       # Start development server
npm run build     # Build for production
npm run start     # Start production server
npm run lint      # Run ESLint
npm test          # Run tests
```

## Architecture

```
flowforge/
├── backend/                 # Go backend
│   ├── cmd/                # Entry points
│   │   ├── api/            # Main API server
│   │   ├── migrate/        # Migration runner
│   │   └── seed/           # Database seeder
│   ├── internal/           # Private packages
│   │   ├── auth/           # JWT & passwords
│   │   ├── config/         # Configuration
│   │   ├── dag/            # DAG validation & topological sort
│   │   ├── db/             # Database connections
│   │   ├── execution/      # Workflow engine
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # Auth, RBAC, rate limiting
│   │   ├── migrate/        # Migration logic
│   │   ├── models/         # Data models
│   │   ├── repository/     # Database access
│   │   ├── scheduler/      # Cron job scheduler
│   │   ├── server/         # Fiber server & routing with graceful shutdown
│   │   └── validator/      # Input validation
│   ├── migrations/         # SQL migrations
│   └── tests/              # Integration tests
├── frontend/               # Next.js frontend
│   ├── src/
│   │   ├── app/            # Next.js App Router
│   │   ├── components/     # React components
│   │   │   ├── nodes/      # React Flow custom nodes
│   │   │   └── __tests__/  # Component tests
│   │   └── lib/            # Utilities & API client
│   │       └── __tests__/  # Unit tests
│   └── public/             # Static assets
├── .github/workflows/      # GitHub Actions CI/CD
├── docker-compose.yml      # Development environment
├── ARCHITECTURE.md         # Architecture documentation
├── TASKS.md                # Task tracking & status
└── DOCKER.md               # Docker deployment guide
```

## Node Types

| Type      | Description                                                     |
| --------- | --------------------------------------------------------------- |
| HTTP      | Make HTTP requests with configurable methods, headers, and URLs |
| Delay     | Pause workflow execution for a specified number of seconds      |
| Script    | Execute custom scripts (placeholder for future implementation)  |
| Condition | Branch workflow execution based on conditional logic            |

## Execution Model

- **Parallel Execution**: Nodes at the same level execute in parallel
- **Timeout Handling**: Workflows automatically fail after timeout
- **Retry Logic**: Failed nodes can be retried with exponential backoff
- **Status Tracking**: Real-time status updates via SSE

## Security

- **Tenant Isolation**: All data scoped to `tenant_id`
- **JWT Authentication**: 24-hour token expiry with HS256 signing
- **RBAC**: Admin, Editor, Viewer roles enforced via middleware
- **Input Validation**: Whitelist validation, length limits, sanitization
- **SQL Injection Prevention**: Parameterized queries via pgx
- **Rate Limiting**: Per-IP and per-user limits (auth: 10/min, read: 100/min, write: 30/min, trigger: 10/min)
- **Webhook Security**: HMAC SHA256 signature verification

## Testing

### Backend

- DAG validation & topological sort tests
- JWT authentication tests (123 tests)
- Input validator tests
- Integration tests (19 suites)
- GitHub Actions CI with coverage threshold

### Frontend

- API client tests
- Auth context tests
- Component tests (dashboard, workflows, permissions)
- UI/UX tests
- GitHub Actions CI with coverage threshold

## License

MIT

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

- [ ] Enhanced HTTP node (auth, retries)
- [ ] Script execution with sandboxing
- [ ] Conditional branching
- [ ] AI-powered failure analysis
- [ ] Workflow templates
- [ ] Export/import workflows
- [ ] Audit log
- [ ] Metrics export (Prometheus, Grafana)
- [ ] Sub-workflows
