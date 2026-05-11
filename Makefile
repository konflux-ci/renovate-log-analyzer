SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

# Defaults for `make run-dev` only. Override on the command line, for example:
#   make run-dev KITE_API_URL=https://kite.example.com LOG_FILE=./path/to/logs.json
# Shell exports of the same names are respected (`?=` does not replace an already-defined variable).

NAMESPACE ?= your-namespace
KITE_API_URL ?= https://kite-api.example.com
KITE_AUTH_TOKEN_FILE ?= ./.local/secrets/kite-token
LOG_FILE ?= ./pkg/doctor/testdata/test_logs.json

.PHONY: help
help:
	@echo "Targets:"
	@echo "  make setup   - one-command dev setup (downloads go modules and creates mock token file)"
	@echo "  make test    - run unit tests"
	@echo "  make run-dev - run analyzer (env vars KITE_API_URL, KITE_AUTH_TOKEN_FILE and LOG_FILE may be overridden, "
	@echo "                e.g. make run-dev KITE_API_URL=... LOG_FILE=...)"
	@echo "                defaults: "
	@echo "                  NAMESPACE defaults to your-namespace"
	@echo "                  KITE_API_URL defaults to https://kite-api.example.com"
	@echo "                  KITE_AUTH_TOKEN_FILE defaults to ./.local/secrets/kite-token"
	@echo "                  LOG_FILE defaults to ./pkg/doctor/testdata/test_logs.json"

.PHONY: setup
setup:
	@command -v go >/dev/null 2>&1 || { echo "go is required (see https://go.dev/dl/)"; exit 1; }
	@echo "Downloading Go modules..."
	@go mod download
	@mkdir -p .local/secrets
	@if [ ! -f .local/secrets/kite-token ]; then \
	  printf "mock-token\n" > .local/secrets/kite-token; \
	  echo "Created .local/secrets/kite-token (mock)"; \
	else \
	  echo ".local/secrets/kite-token already exists; leaving as-is"; \
	fi
	@echo "Setup complete. Run one of the following commands:"
	@echo "- make test"
	@echo "- make run-dev"

.PHONY: test
test:
	go test ./...

.PHONY: run-dev
run-dev:
	@NAMESPACE="$(NAMESPACE)" \
	KITE_API_URL="$(KITE_API_URL)" \
	KITE_AUTH_TOKEN_FILE="$(KITE_AUTH_TOKEN_FILE)" \
	GIT_HOST=github.com \
	REPOSITORY=owner/repo \
	BRANCH=main \
	LOG_FILE="$(LOG_FILE)" \
	go run ./cmd/log-analyzer/main.go --dev