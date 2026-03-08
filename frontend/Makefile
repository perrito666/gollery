SHELL := /bin/bash

.PHONY: help build dev clean install resolve-theme print-effective-theme test

help:
	@echo "Targets:"
	@echo "  make install               - install npm dependencies"
	@echo "  make build                 - bundle for production"
	@echo "  make dev                   - start dev server with watch"
	@echo "  make resolve-theme         - resolve site overrides"
	@echo "  make print-effective-theme - show which views/styles are active"
	@echo "  make test                  - run unit tests"
	@echo "  make clean                 - remove dist artifacts"

install:
	npm install

resolve-theme: install
	node scripts/resolve-theme.js

print-effective-theme: install
	node scripts/resolve-theme.js --print

build: resolve-theme
	mkdir -p dist/styles
	npx esbuild src/main.js --bundle --outfile=dist/app.js --format=esm --minify
	cp public/index.html dist/
	cp src/ui-default/styles/main.css dist/styles/main.css
	@# Layer site CSS on top if it exists
	@if [ -f src/site/styles/main.css ]; then cat src/site/styles/main.css >> dist/styles/main.css; fi

test: install
	npm test

dev: resolve-theme
	npx esbuild src/main.js --bundle --outfile=public/app.js --format=esm --watch --servedir=public

clean:
	rm -rf dist node_modules src/_resolved
