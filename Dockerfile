# ============================================
# Stage 1: Build Frontend
# ============================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY apps/railzway/package*.json ./
RUN npm ci

# Copy frontend source and build
COPY apps/railzway/ ./
RUN npm run build

# ============================================
# Stage 2: Build Backend
# ============================================
FROM golang:1.25-alpine AS backend-builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /workspace/railzway-cloud ./apps/railzway/main.go

# ============================================
# Stage 3: Runtime
# ============================================
FROM gcr.io/distroless/static-debian12 AS runtime

# Copy backend binary
COPY --from=backend-builder /workspace/railzway-cloud /usr/local/bin/railzway-cloud

# Copy frontend build artifacts
COPY --from=frontend-builder /app/frontend/dist /app/dist

# Copy SQL migrations
COPY sql/migrations /app/sql/migrations

EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/railzway-cloud"]
