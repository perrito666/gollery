SHELL := /bin/bash

REPO_NAME := gollery
DIST_DIR := dist
PACKAGE := $(REPO_NAME)-starter-kit-full.zip

.PHONY: help tree docs prompts package clean backend-build backend-test

help:
	@echo "Targets:"
	@echo "  make tree            - print starter tree"
	@echo "  make docs            - list design docs"
	@echo "  make prompts         - list agent prompts"
	@echo "  make package         - create starter zip"
	@echo "  make clean           - remove dist artifacts"
	@echo "  make backend-build  - compile backend"
	@echo "  make backend-test   - run backend tests"

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

clean:
	@rm -rf $(DIST_DIR)
