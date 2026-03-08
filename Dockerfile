# Stage 1: Build Go backend
FROM golang:1.25-bookworm AS backend-builder
WORKDIR /src
COPY backend/ ./backend/
WORKDIR /src/backend
RUN go mod download
RUN CGO_ENABLED=0 go build -o /galleryd ./cmd/galleryd

# Stage 2: Build frontend
FROM node:22-bookworm-slim AS frontend-builder
RUN apt-get update && apt-get install -y --no-install-recommends make && rm -rf /var/lib/apt/lists/*
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN make build

# Stage 3: Runtime image
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/*

COPY --from=backend-builder /galleryd /usr/local/bin/galleryd
COPY --from=frontend-builder /src/frontend/dist /var/lib/gollery/frontend

# Default config
COPY <<'EOF' /etc/gollery/gollery.json
{
  "content_root": "/data/content",
  "cache_dir": "/data/cache",
  "listen_addr": ":8080"
}
EOF

EXPOSE 8080

ENTRYPOINT ["galleryd"]
CMD ["--config", "/etc/gollery/gollery.json"]
