# Renovate Log Analyzer

A Go-based tool for analyzing [Renovate](https://github.com/renovatebot/renovate) logs and reporting issues to [Kite API](https://github.com/Issues-Dashboard/kite). Part of the [MintMaker](https://github.com/konflux-ci/mintmaker) ecosystem.

## Overview

This tool runs as the final step of Tekton pipelines created by the [MintMaker controller](https://github.com/konflux-ci/mintmaker).

It reads Renovate JSON logs, extracts and categorizes issues, and reports them to Kite for the Konflux UI Issues dashboard.

### Key Features

- **Dual Processing**: Level-based (ERROR/FATAL) and message-based pattern matching
- **Smart Error Extraction**: Condenses verbose stack traces to essential information
- **Categorization**: Automatically categorizes issues as Errors, Warnings, or Infos
- **Kite Integration**: Sends webhook notifications to Kite API with analyzed results

## How It Works

1. **Log Processing**: reads JSON logs; extracts ERROR (50) and FATAL (60) entries
2. **Aggregation**: groups level-based errors by message with duplicate counts
3. **Selectors**: pattern-matches messages into actionable issues
4. **Health Check**: probes Kite before sending webhooks
5. **Webhooks**: sends success, failure, or custom payloads based on findings

## Token File

The tool reads the token from **`KITE_AUTH_TOKEN_FILE`** (path on the host or inside the container).

For a **local mock** token and modules, run **`make setup`** (see [Quick start](#quick-start)). That creates `.local/secrets/kite-token` only if it is missing.

> **Note:** With a mock token and placeholder `KITE_API_URL`, log analysis still runs; the HTTP step to Kite may fail. That is expected.

If real token is available, set `KITE_AUTH_TOKEN_FILE` to the real secret path. When using a real token, be sure to change the `NAMESPACE` env var to the correct one as well.

## Quick Start

```bash
make setup
```

Smoke run (placeholder Kite URL; webhook step may fail as noted under [Token File](#token-file)):

```bash
make run-dev
```

Overrides (use real values when available):

```bash
make run-dev \
  KITE_API_URL=https://kite.example.com \
  KITE_AUTH_TOKEN_FILE=/path/to/your/token \
  LOG_FILE=./path/to/renovate-logs.json
```

Shell-first workflow (real backend): export the same variable names, then `go run ./cmd/log-analyzer/main.go --dev`. See [Local testing](docs/README.md#local-testing).

## Run From Source

Use the root [`Makefile`](Makefile): `make help`, `make setup`, `make test`, `make run-dev`. Extended notes: [Makefile (local development)](docs/README.md#makefile-local-development).

Or follow [docs/README.md](docs/README.md) for architecture and [local testing](docs/README.md#local-testing).

## Run With Container Image

Image: [`quay.io/konflux-ci/renovate-log-analyzer:latest`](https://quay.io/konflux-ci/renovate-log-analyzer:latest).

```bash
podman run --rm \
  -e NAMESPACE="your-namespace" \
  -e KITE_API_URL="https://kite-api.example.com" \
  -e KITE_AUTH_TOKEN_FILE="/run/secrets/kite-token" \
  -e LOG_FILE="/work/renovate-logs.json" \
  -v "./pkg/doctor/testdata/fatal_exit_logs.json:/work/renovate-logs.json:ro" \
  -v "./.local/secrets/kite-token:/run/secrets/kite-token:ro" \
  quay.io/konflux-ci/renovate-log-analyzer:latest
```

## Documentation

[docs/README.md](docs/README.md) — architecture, selectors, `extractUsefulError`, local testing, Makefile, Kite client.

## Environment Variables

### Required

- **`NAMESPACE`**: Kubernetes namespace
- **`KITE_API_URL`**: Kite API base URL
- **`KITE_AUTH_TOKEN_FILE`**: path to a file containing the auth token

### Optional

- **`GIT_HOST`**: default `unknown`
- **`REPOSITORY`**: default `unknown`
- **`BRANCH`**: default `unknown`
- **`LOG_FILE`**: default `/workspace/shared-data/renovate-logs.json`
- **`PIPELINE_RUN`**: default `unknown`

### Flags

- **`--dev`**: verbose logging and source locations

## Project Structure

```text
renovate-log-analyzer/
├── Makefile                 # Local dev: make help, setup, test, run-dev
├── cmd/
│   └── log-analyzer/
│       └── main.go          # Entry point
├── pkg/
│   ├── doctor/              # Log analysis package
│   │   ├── checks.go        # Selector definitions
│   │   ├── models.go        # Data models
│   │   ├── report.go        # Report generation
│   │   └── log_reader.go    # Log processing
│   └── kite/                # Kite API client
│       └── client.go
└── docs/
    └── README.md            # Detailed documentation
```

## License

Licensed under the Apache License, Version 2.0. See [licenses/LICENSE](licenses/LICENSE) for details.
