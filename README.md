# FlowForge

Real-time multi-tenant workflow orchestration platform inspired by Zapier + GitHub Actions.

## Features

- **Visual Workflow Builder**: Drag-and-drop DAG editor using React Flow
- **Real-time Execution**: Live monitoring with Server-Sent Events (SSE)
- **Multi-tenant**: Complete tenant isolation
- **Role-based Access Control**: Admin, Editor, Viewer roles
- **Workflow Types**: HTTP requests, delays, scripts, conditions
- **Execution Engine**: Parallel processing with topological sorting
- **Version Control**: Automatic workflow versioning

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
- **Styling**: Tailwind CSS
- **DAG Editor**: React Flow
- **State**: React Context API

## Quick Start

### Prerequisites

- Go 1.26+
- Node.js 20+
- Docker & Docker Compose (optional)
- PostgreSQL 16+
- Redis 7+

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/lavianrose/flowforge.git
cd flowforge

# Start all services
docker-compose up -d

# Run database migrations
docker-compose exec backend make migrate-up

# Seed admin user
docker-compose exec backend make seed

# Access the application
# Frontend: http://localhost:3001
# Backend API: http://localhost:3000
```

### Manual Setup

#### Backend

```bash
cd backend

# Install dependencies
go mod download

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Start PostgreSQL & Redis (Docker)
docker-compose up -d postgres redis

# Run migrations
make migrate-up

# Seed admin user
make seed

# Start server
make run
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Configure environment
cp .env.local.example .env.local
# Edit .env.local with your API URL

# Start development server
npm run dev
```

## Default Credentials

- **Email**: admin@flowforge.local
- **Password**: admin123

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
GET /api/v1/workflows
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

### Runs

#### List Runs
```http
GET /api/v1/runs
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

## Development

### Backend Commands

```bash
make build        # Build binaries
make run          # Start server
make migrate-up   # Run migrations
make migrate-down # Rollback migrations
make seed         # Seed admin user
make clean        # Clean build artifacts
```

### Frontend Commands

```bash
npm run dev       # Start development server
npm run build     # Build for production
npm run start     # Start production server
npm run lint      # Run ESLint
```

## Architecture

```
flowforge/
├── backend/              # Go backend
│   ├── cmd/             # Entry points
│   │   ├── api/         # Main API server
│   │   ├── migrate/     # Migration runner
│   │   └── seed/        # Database seeder
│   ├── internal/        # Private packages
│   │   ├── auth/        # JWT & passwords
│   │   ├── config/      # Configuration
│   │   ├── dag/         # DAG validation
│   │   ├── db/          # Database connections
│   │   ├── execution/   # Workflow engine
│   │   ├── handlers/    # HTTP handlers
│   │   ├── middleware/  # Fiber middleware
│   │   ├── migrate/     # Migration logic
│   │   ├── models/      # Data models
│   │   └── repository/  # Database access
│   └── migrations/      # SQL migrations
├── frontend/            # Next.js frontend
│   ├── src/
│   │   ├── app/         # Next.js App Router
│   │   ├── components/  # React components
│   │   └── lib/         # Utilities & API client
│   └── public/          # Static assets
└── docker-compose.yml   # Development environment
```

## Node Types

### HTTP Request
Make HTTP requests with configurable methods, headers, and URLs.

### Delay
Pause workflow execution for a specified number of seconds.

### Script
Execute custom scripts (placeholder for future implementation).

### Condition
Branch workflow execution based on conditional logic (placeholder for future implementation).

## Execution Model

- **Parallel Execution**: Nodes at the same level execute in parallel
- **Timeout Handling**: Workflows automatically fail after timeout
- **Retry Logic**: Failed nodes can be retried with exponential backoff
- **Status Tracking**: Real-time status updates via SSE

## Security

- **Tenant Isolation**: All data scoped to tenant_id
- **JWT Authentication**: 24-hour token expiry
- **Role-based Access**: Admin, Editor, Viewer roles
- **Input Validation**: All inputs validated before execution
- **SQL Injection**: Using parameterized queries (pgx)

## License

MIT

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

- [ ] Cron scheduling
- [ ] Webhook triggers
- [ ] Enhanced HTTP node (auth, retries)
- [ ] Script execution with sandboxing
- [ ] Conditional branching
- [ ] AI-powered failure analysis
- [ ] Workflow templates
- [ ] Export/import workflows
