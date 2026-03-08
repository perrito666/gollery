SHELL := /bin/bash

REPO_NAME := gollery
DIST_DIR := dist
PACKAGE := $(REPO_NAME)-starter-kit-full.zip

.PHONY: help tree docs prompts package clean
.PHONY: backend-build backend-test frontend-build frontend-test
.PHONY: build test verify status docker-build docker-up docker-down

help:
	@echo "Targets:"
	@echo "  make build            - compile backend + bundle frontend"
	@echo "  make test             - run all tests"
	@echo "  make verify           - build + test + vet (full verification)"
	@echo "  make status           - show current prompt and project state"
	@echo ""
	@echo "  make backend-build    - compile backend"
	@echo "  make backend-test     - run backend tests"
	@echo "  make frontend-build   - bundle frontend"
	@echo "  make frontend-test    - run frontend tests (when available)"
	@echo ""
	@echo "  make tree             - print starter tree"
	@echo "  make docs             - list design docs"
	@echo "  make prompts          - list agent prompts"
	@echo "  make package          - create starter zip"
	@echo "  make docker-build     - build Docker image"
	@echo "  make docker-up        - start with docker compose"
	@echo "  make docker-down      - stop docker compose"
	@echo ""
	@echo "  make clean            - remove dist artifacts"

build: backend-build frontend-build

test: backend-test frontend-test

verify: build
	$(MAKE) -C backend vet
	$(MAKE) -C backend test
	@echo "--- verification complete ---"

status:
	@echo "=== gollery project status ==="
	@echo ""
	@echo "Phase 1 (01-18): COMPLETE"
	@echo "Next prompt: 19 (frontend shared utilities)"
	@echo "Full plan: docs/agent-workflow.md"
	@echo ""
	@echo "Backend tests:"
	@cd backend && go test ./... 2>&1 | tail -1
	@echo ""
	@echo "Frontend bundle:"
	@cd frontend && npx esbuild src/main.js --bundle --outfile=/dev/null --format=esm 2>&1 || echo "  (run 'make frontend-build' first)"

tree:
	@find . -maxdepth 4 | sort

docs:
	@find docs -maxdepth 2 -type f | sort

prompts:
	@find docs/agent-prompts -maxdepth 1 -type f | sort

package:
	@mkdir -p $(DIST_DIR)
	@zip -qr $(DIST_DIR)/$(PACKAGE) . -x "dist/*" -x ".git/*"

backend-build:
	$(MAKE) -C backend build

backend-test:
	$(MAKE) -C backend test

frontend-build:
	$(MAKE) -C frontend build

frontend-test:
	@if [ -f frontend/Makefile ] && grep -q '^test:' frontend/Makefile 2>/dev/null; then \
		$(MAKE) -C frontend test; \
	else \
		echo "frontend tests not configured yet (prompt 45)"; \
	fi

docker-build:
	docker build -t gollery .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

clean:
	@rm -rf $(DIST_DIR)
	$(MAKE) -C backend clean 2>/dev/null || true
	$(MAKE) -C frontend clean 2>/dev/null || true
