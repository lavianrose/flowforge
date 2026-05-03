# Docker Setup Guide

This guide explains how to run FlowForge using Docker Compose in different scenarios.

## Quick Start (Everything in Docker)

```bash
# Copy and configure environment
cp .env.example .env

# Start all services
docker-compose up -d

# Access the application
# Frontend: http://127.0.0.1:3001
# Backend API: http://127.0.0.1:3000
```

## Configuration Scenarios

### Scenario 1: All Services in Docker, Accessed from Browser (Recommended)

**Use case:** Running everything in Docker, accessed from host browser

**`.env` configuration:**
```env
NEXT_PUBLIC_API_URL=http://127.0.0.1:3000/api/v1
```

**How it works:**
- Frontend runs in Docker container
- Backend runs in Docker container
- **Important:** Browser accesses frontend, not Docker services
- Browser cannot resolve Docker service names like `backend`
- Must use `127.0.0.1` so browser can connect via port mapping
- Frontend code makes API calls to `http://127.0.0.1:3000`
- Docker port mapping: `127.0.0.1:3000 → backend:3000`

**Access URLs:**
- Frontend: http://127.0.0.1:3001 (in browser)
- Backend API: http://127.0.0.1:3000 (for browser/external access)
- Backend internal: http://backend:3000 (for Docker service-to-service)

**⚠️ Common Mistake:**
Using `NEXT_PUBLIC_API_URL=http://backend:3000/api/v1` will cause `ERR_NAME_NOT_RESOLVED`
because browsers cannot resolve Docker internal service names.

### Scenario 2: Frontend Local, Backend in Docker

**Use case:** Developing frontend locally with backend containerized

**`.env` configuration:**
```env
NEXT_PUBLIC_API_URL=http://127.0.0.1:3000/api/v1
```

**Steps:**
```bash
# Terminal 1: Start backend services only
docker-compose up postgres redis backend -d

# Terminal 2: Run frontend locally
cd frontend
npm install
npm run dev
```

**Access URLs:**
- Frontend: http://localhost:3001 (or configured Next.js dev port)
- Backend: http://127.0.0.1:3000

### Scenario 3: Everything Local

**Use case:** Full local development

**`.env` configuration:**
```env
NEXT_PUBLIC_API_URL=http://localhost:3000/api/v1
```

**Steps:**
```bash
# Terminal 1: Start databases
docker-compose up postgres redis -d

# Terminal 2: Run backend
cd backend
go run cmd/api/main.go

# Terminal 3: Run frontend
cd frontend
npm run dev
```

**Access URLs:**
- Frontend: http://localhost:3001 (Next.js dev server)
- Backend: http://localhost:3000

## Cross-Platform Compatibility

### Why 127.0.0.1 instead of localhost?

- **127.0.0.1** is an IP address that works consistently across all operating systems
- **localhost** resolution can vary:
  - Linux/macOS: Usually resolves to 127.0.0.1
  - Windows: May resolve to IPv6 ::1 first
  - Some systems: May not resolve at all without network configuration
- Docker networking: 127.0.0.1 is more reliable for port mapping

### Port Bindings

The `docker-compose.yml` uses explicit `127.0.0.1` binding:

```yaml
ports:
  - "127.0.0.1:3000:3000"  # Backend
  - "127.0.0.1:3001:3001"  # Frontend
```

This ensures services are only accessible from the host machine, not from external networks.

## Environment Variables

### Required Variables

```env
# Backend
JWT_SECRET=your-secret-key-here
PORT=3000
ENV=production

# Database
POSTGRES_URL=postgres://flowforge:flowforge@postgres:5432/flowforge?sslmode=disable
REDIS_URL=redis:6379

# Frontend
NEXT_PUBLIC_API_URL=http://backend:3000/api/v1
```

### Variable Precedence

1. **Runtime environment** (highest priority)
2. **docker-compose.yml** `environment` section
3. **`.env`** file
4. **Default value** in docker-compose (`${VAR:-default}`)

## Troubleshooting

### Frontend cannot connect to backend

**Problem:** API calls failing from browser

**Solutions:**

1. **Check NEXT_PUBLIC_API_URL in browser console:**
   ```javascript
   console.log(process.env.NEXT_PUBLIC_API_URL)
   ```

2. **Verify backend is accessible:**
   ```bash
   curl http://127.0.0.1:3000/api/v1/health
   ```

3. **Check Docker logs:**
   ```bash
   docker-compose logs backend
   docker-compose logs frontend
   ```

4. **Verify port bindings:**
   ```bash
   docker-compose ps
   ```

### "Connection refused" errors

**Problem:** Services cannot communicate

**Solutions:**

1. **Ensure all services are running:**
   ```bash
   docker-compose ps
   ```

2. **Check service health:**
   ```bash
   docker-compose ps postgres
   docker-compose ps redis
   ```

3. **Verify network:**
   ```bash
   docker network inspect flowforge_default
   ```

### Changes not appearing

**Problem:** Code changes not reflected

**Solutions:**

1. **Rebuild containers:**
   ```bash
   docker-compose up -d --build
   ```

2. **Clear cache and rebuild:**
   ```bash
   docker-compose down -v
   docker-compose up -d --build
   ```

## Default Credentials

After starting the services, default admin credentials are created:

- **Email:** admin@flowforge.local
- **Password:** admin123

**Important:** Change these credentials immediately after first login in production!

## Database Seeding

The backend automatically runs migrations and seeds on startup:

```bash
# View seeding logs
docker-compose logs backend | grep seed

# Re-run seeding (if needed)
docker-compose exec backend ./seed
```

## Production Deployment

For production, ensure:

1. ✅ Change `JWT_SECRET` to a strong random value
2. ✅ Use strong database passwords
3. ✅ Enable SSL/TLS for external connections
4. ✅ Set up proper backup strategy
5. ✅ Configure resource limits in docker-compose.yml
6. ✅ Use Docker secrets or vault for sensitive data
7. ✅ Set `ENV=production`
8. ✅ Remove or change default credentials

## Useful Commands

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# View logs
docker-compose logs -f

# Restart specific service
docker-compose restart backend

# Execute command in container
docker-compose exec backend sh
docker-compose exec postgres psql -U flowforge -d flowforge

# Rebuild specific service
docker-compose up -d --build backend

# Remove everything (including volumes)
docker-compose down -v

# Check resource usage
docker stats
```
