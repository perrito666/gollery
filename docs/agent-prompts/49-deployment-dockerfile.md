# Prompt 49 — Deployment Dockerfile

Create a multi-stage Dockerfile for the gollery monorepo.

Implement:
- stage 1: build the Go backend binary (`galleryd`)
- stage 2: build the frontend (`make build` in frontend/)
- stage 3: minimal runtime image (e.g., `debian:bookworm-slim` or `distroless`) with:
  - the `galleryd` binary
  - the frontend `dist/` directory
  - a default `gollery.json` config
- expose port 8080
- set `ENTRYPOINT` to `galleryd`
- add a `.dockerignore` at the repo root

Do not add docker-compose yet.
Do not include PostgreSQL in the image.
