SHELL := /bin/bash

BINARY := galleryd
USERS_BINARY := gollery-users
CMD_DIR := ./cmd/galleryd
USERS_CMD_DIR := ./cmd/gollery-users

.PHONY: help build test vet clean

help:
	@echo "Targets:"
	@echo "  make build        - compile galleryd and gollery-users"
	@echo "  make test         - run all tests"
	@echo "  make vet          - run go vet"
	@echo "  make clean        - remove build artifacts"

build:
	go build -o $(BINARY) $(CMD_DIR)
	go build -o $(USERS_BINARY) $(USERS_CMD_DIR)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY) $(USERS_BINARY)
