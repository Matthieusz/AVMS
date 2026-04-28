# syntax=docker/dockerfile:1

# -------------------------------
# Stage 1: Build frontend
# -------------------------------
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN corepack enable && corepack prepare pnpm@latest --activate && pnpm install --frozen-lockfile
COPY frontend/ .
RUN pnpm run build

# -------------------------------
# Stage 2: Build Go backend
# -------------------------------
FROM golang:1.24-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download
COPY . .
# Copy built frontend into the expected static directory
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o avms-api ./cmd/api

# -------------------------------
# Stage 3: Final runtime image
# -------------------------------
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend-builder /app/avms-api .
COPY --from=backend-builder /app/frontend/dist ./frontend/dist
COPY --from=backend-builder /app/internal/database/migrations ./internal/database/migrations

ENV AVMS_STATIC_DIR=/app/frontend/dist
ENV AVMS_PORT=8080
ENV GIN_MODE=release
ENV AVMS_LOG_FORMAT=json

EXPOSE 8080

ENTRYPOINT ["./avms-api"]
