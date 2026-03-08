# Prompt 50 — Deployment docker-compose

Create a docker-compose setup for local development and testing.

Implement:
- `docker-compose.yml` at the repo root with:
  - `galleryd` service built from the Dockerfile
  - `postgres` service using `postgres:16-alpine`
  - a shared volume for the content root
  - environment variables for PostgreSQL DSN and config
- a sample `gollery.json` config file at the repo root
- a sample `users.json` with one admin user (bcrypt-hashed default password)
- update the root `Makefile` with `docker-build` and `docker-up` targets

Do not include production deployment configs (Kubernetes, Ansible, etc.).
