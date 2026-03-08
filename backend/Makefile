SHELL := /bin/bash

BINARY := galleryd
CMD_DIR := ./cmd/galleryd

.PHONY: help build test vet clean

help:
	@echo "Targets:"
	@echo "  make build  - compile galleryd"
	@echo "  make test   - run all tests"
	@echo "  make vet    - run go vet"
	@echo "  make clean  - remove build artifacts"

build:
	go build -o $(BINARY) $(CMD_DIR)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
