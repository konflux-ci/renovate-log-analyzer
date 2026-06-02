---
name: add-selector
description: >-
  Add Renovate log selector in pkg/doctor (checks.go). Use when adding message-based pattern matching, log categorization, new Renovate msg patterns, or Issues dashboard findings from Renovate JSON logs.
---

# Add Renovate log selector

Guide for adding a **message-based selector** in `pkg/doctor`. Selectors feed `SimpleReport` and the `mintmaker-custom` Kite webhook. They are separate from **level-based** ERROR/FATAL handling (`pipeline-failure`).

## Before you start

1. Get a real Renovate log line that should trigger the check either from the user or from analyzing a log file (if provided). When you do not have the whole json log, **do not** attempt to create a selector.
2. Confirm the exact `msg` substring Renovate emits (copy from JSON, do not guess).
3. Decide severity: `report.Error`, `report.Warning`, or `report.Info`.

## How matching works

- Registered in `init()` via `registerSelector(substring, checkFunc)` in `pkg/doctor/checks.go`.
- On each parsed line, `log_reader.go` runs `strings.Contains(entry.Msg, selector)`.
- The selector string must appear **verbatim** as a substring of `msg` (not regex).
- Selectors run on **all** log levels.

## Implementation checklist

Copy an existing check in `checks.go` (`prLimitReached`, `renovateConfigErrors`, `rawExecError`, `platformCommitError`) and adapt.

- [ ] **1. Check function** — `func myCheck(line *LogEntry, report *SimpleReport)` (use descriptive function name) in `checks.go`:
  - Read fields from `line.Extras` (and `line.Msg` if needed). Unknown JSON keys land in `Extras` after `parseLogLine`. Analyze the real json log and extract a useful fields based on that.
  - Call `report.Error` / `report.Warning` / `report.Info` with optional key/value pairs: `report.Error("Title", "Branch", line.Extras["branch"], "Message", text)`.
  - For long `err.message` strings, use `extractUsefulErrorDefault(message)` (see `rawExecError`).
  - Return early if required structured data is missing (see `rawExecError`, `platformCommitError`).
- [ ] **2. Register** — In `init()`, add `registerSelector("<exact substring from Renovate msg>", myCheck)`.
- [ ] **3. Test fixture** — Append **one** minimal JSON line to `pkg/doctor/testdata/test_logs.json` (or a dedicated file under `testdata/` if the line is huge). Requirements:
  - Valid NDJSON: one JSON object per line.
  - `msg` contains the registered selector substring.
  - **Synthetic / anonymized only**: no real tokens, cookies, passwords, or internal-only URLs. Use placeholders (`example-org`, `example.com`, `**redacted**`). `gitleaks` runs on pre-commit.
  - Mirror structure from neighboring lines in `test_logs.json` (`level`, `branch`, `err`, etc.).
- [ ] **4. Tests** — Add table-driven `*_test.go` in `pkg/doctor` when behavior is non-trivial: run `ProcessLogFile` on a small fixture and assert `SimpleReport` slices. No live HTTP.
- [ ] **5. Documentation** — Update selector list in `docs/README.md`. Update root `README.md` only if user-facing behavior changes.
- [ ] **6. Verify**

```bash
go fmt ./pkg/doctor/...
golangci-lint run ./pkg/doctor
make test
make run-dev LOG_FILE=./pkg/doctor/testdata/test_logs.json
```

Use `go run ./cmd/log-analyzer/main.go --dev` with the same env vars as `make run-dev` when you need custom `GIT_HOST` / `REPOSITORY` / `BRANCH`. With a mock Kite URL, parsing and `--dev` console output still work; webhook HTTP may fail (expected).

## Report formatting

- `formatSimpleMessage` builds strings from title + alternating key/value fields.
- `Message` values are appended on a new line; other fields use `| key: value`.

## Common mistakes

- **Wrong selector string** — Must match Renovate `msg`, not log text inside `err.message` only.
- **Regex in selector** — Not supported; use exact substring or handle parsing inside the check function.
- **Bloated fixtures** — Do not paste full production stack traces; trim to fields your check reads.
- **Skipping docs** — Selector list in `docs/README.md` must stay in sync.
- **Changing Kite payloads** — Selector work stays in `pkg/doctor`. Webhook schema changes belong in `pkg/kite/` and require alignment with the [Kite backend](https://github.com/Issues-Dashboard/kite/tree/main/packages/backend).

## References

- Architecture and selector pattern: `docs/README.md` (Selector Pattern, Selector List, Local Testing).
- Code: `pkg/doctor/checks.go`, `pkg/doctor/log_reader.go`, `pkg/doctor/report.go`, `pkg/doctor/models.go`.
- Repo conventions: `AGENTS.md`.
