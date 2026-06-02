# AGENTS.md

## Project overview

Renovate log analyzer: analyzes structured logs from Renovate in JSON format, reports findings to Kite. It is meant to be run as a last step of MintMaker Tekton pipeline right after the Renovate step.

Stack: Go. Docs: root `README.md`, `docs/README.md`.

## Architecture

- `cmd/log-analyzer/`: CLI—env, `slog`, `doctor.ProcessLogFile`, Kite webhooks.
- `pkg/doctor/`: Parse lines, ERROR/FATAL aggregation, selector checks (`Selectors` + `init()` in `checks.go`), `SimpleReport`.
- `pkg/kite/`: HTTP client for health + webhooks.
- `pkg/doctor/testdata/`: Sample log files for local runs.

## Commands

- **Setup**: `make setup` — modules + mock token under `.local/secrets/`.
- **Test**: `make test` — runs `go test ./...`.
- **Smoke run**: `make run-dev` (override `LOG_FILE`, `KITE_*`, `NAMESPACE` as needed - explained in `docs/README.md`).
- **Quick lint:** Packages - `golangci-lint run ./pkg/doctor` (swap for `./pkg/kite`, `./cmd/log-analyzer`, etc.). Single file - `go fmt ./pkg/doctor/log_reader.go` (insert path to edited file).
- **Quick type-check / compile:** `go build -o /dev/null ./pkg/doctor` — same package path idea; no separate `mypy`-style tool in Go.

## Conventions

- **Edits**: Match existing Apache 2.0 headers, `slog` usage, and error wrapping (`fmt.Errorf` with `%w`). Follow Go naming (exported: PascalCase; unexported: camelCase).
- **Formatting**: Keep Go gofmt clean (`go fmt ./...`); for Markdown, respect `markdownlint` (see `.markdownlint.json`).
- **Selectors**: See [Pattern References](#pattern-references). Add a mock log line in `pkg/doctor/testdata/test_logs.json` to test the new selector.
- **When adding new behavior**: Add tests, prefer table-driven `*_test.go` in the same package. No live HTTP in tests (fake `kite` at call sites).
- **External libraries**: Use external libraries only when completely necessary. Run `go mod tidy` whenever dependencies are added, removed, or upgraded. Include resulting `go.mod` and `go.sum` in the commit.
- **Secrets**: Never commit tokens; `gitleaks` runs in pre-commit.
- **Sample logs**: Files under `pkg/doctor/testdata/` have to be synthetic or anonymized: no real tokens, cookies, or internal-only URLs.
- **Documentation**: Update the corresponding parts of documentation (root `README.md` and `docs/README.md`) to reflect changes made when applicable (env vars, flags, selectors, pipeline behavior, Kite payloads).
- **Lint check**: After changes, run `golangci-lint` on the edited package. Fix any issues.
- **Kite package**: Changes in `pkg/kite/` must stay consistent with the official Kite service and its repository. Verify against the Kite repo (or maintainers) before merging behavior changes.
- **Commits**: Run `pre-commit install` before commiting. Use conventional commits (e.g., `feat:`, `fix:`, `chore:`).

## Pattern References

- **New selector**: Use the `add-selector` skill (`skills/add-selector/SKILL.md`). See also `docs/README.md` and `pkg/doctor/checks.go`.
