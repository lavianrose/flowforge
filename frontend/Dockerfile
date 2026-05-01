# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies with cache mount for faster rebuilds
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline --no-audit --no-fund

# Copy source code and config
COPY next.config.ts ./
COPY tsconfig.json ./
COPY .npmrc ./

# Copy public assets
COPY public ./public

# Copy source directory
COPY src ./src

# Build Next.js app with standalone output
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# Runtime stage - minimal alpine
FROM node:20-alpine

WORKDIR /app

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# Copy package files
COPY package*.json ./

# Install only production dependencies
RUN --mount=type=cache,target=/root/.npm \
    npm ci --only=production --prefer-offline --no-audit --no-fund

# Copy built assets from builder
COPY --from=builder --chown=appuser:appuser /app/.next/standalone ./
COPY --from=builder --chown=appuser:appuser /app/.next/static ./.next/static
COPY --from=builder --chown=appuser:appuser /app/public ./public

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3001

# Set environment
ENV NODE_ENV=production
ENV PORT=3001
ENV NEXT_TELEMETRY_DISABLED=1

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001 || exit 1

# Start the application
CMD ["node", "server.js"]
