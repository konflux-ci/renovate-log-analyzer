# Renovate Log Analyzer (used as part of the [MintMaker](https://github.com/konflux-ci/mintmaker) service)

This repository contains a Go implementation for analyzing Renovate logs. The implementation provides level-based error (and fatal) extraction.

Another part of this repo is the KITE client, which in the event of a failure in Renovate sends the extracted errors to the [KITE API](https://github.com/konflux-ci/kite) to be displayed on an Issue dashboard in Konflux UI.

This service is meant to run as last step of `tekton pipeline` created by the [MintMaker controller](https://github.com/konflux-ci/mintmaker).

## Log analyzer

- **`models.go`**: Data models (`LogEntry` and `PodDetails`)
- **`log_reader.go`**: Log processing logic for extracting logs from a `json` file and parsing them into `Go` object as well as the analyzing logic

## Log Levels

Following [Renovate documentation](https://docs.renovatebot.com/troubleshooting/):

- **TRACE**: 10
- **DEBUG**: 20
- **INFO**: 30
- **WARN**: 40
- **ERROR**: 50
- **FATAL**: 60

The log-analyzer looks for errors with level 50 (ERROR) and 60 (FATAL) and extracts the most useful information from them by looking through all the possible fields in their structure. It aggregates duplicate errors and formats them into a summary message.

## KITE client
- **`client.go`**: Contains everything needed to communicate with the [KITE API backend](https://github.com/konflux-ci/kite/tree/main/packages/backend) - defines Payload structures, initializes the client and contains functions to send requests. The client checks KITE API health status and sends webhooks for pipeline success or failure.

## Local Testing

To test the log analyzer locally using `go run ./cmd/log-analyzer/main.go` the following set up is needed:

### Required Environment Variables

The application requires the following environment variables:

- **`NAMESPACE`**: Kubernetes namespace (required)
- **`KITE_API_URL`**: URL to the KITE API endpoint (required)
- **`GIT_HOST`**: Git host (e.g., github.com) (required)
- **`REPOSITORY`**: Repository name (required, underscores will be replaced with slashes)
- **`BRANCH`**: Branch name (required)
- **`LOG_FILE`**: Path to the Renovate log file (optional, defaults to `/workspace/shared-data/renovate-logs.json`)
- **`PIPELINE_RUN`**: Pipeline run identifier (optional, defaults to "unknown")

### Test Log File Format

The log file should contain Renovate JSON logs, with each line being a separate JSON object. Example:

```json
{"level": 20, "msg": "rawExec err", "err": {"message": "Command failed: npm install"}, "branch": "main"}
{"level": 40, "msg": "Reached PR limit - skipping PR creation"}
{"level": 30, "msg": "branches info extended", "branchesInformation": [...]}
{"level": 50, "msg": "Base branch does not exist - skipping", "baseBranch": "feature/old"}
{"level": 60, "msg": "Fatal error occurred", "err": {"message": "Critical failure"}}
```

### Example Test Command

```bash
# Set required environment variables
export NAMESPACE=namespace-name
export KITE_API_URL=https://kite-api.example.com    # or placeholder for testing
export GIT_HOST=github.com
export REPOSITORY=owner/repo                        # or owner_repo (underscores converted to slashes)
export BRANCH=main
export LOG_FILE="./fatal_exit_logs.json"            # path to test log file (optional)
export PIPELINE_RUN=test-run-123                    # optional

# Run the application
go run ./cmd/log-analyzer/main.go
```

### How It Works

1. **Log Processing**: The application reads the log file (default: `/workspace/shared-data/renovate-logs.json`) and extracts ERROR (level 50) and FATAL (level 60) entries.

2. **Error Aggregation**: Errors are aggregated by message, with duplicate counts tracked.

3. **KITE API Health Check**: Before sending webhooks, the application checks the KITE API health status.

4. **Webhook Notification**: 
   - If no errors are found, sends a `pipeline-success` webhook
   - If errors are found, sends a `pipeline-failure` webhook with the aggregated failure reason

5. **Pipeline Identifier**: The pipeline identifier is constructed as `{GIT_HOST}/{REPOSITORY}/{BRANCH}`.

### Notes

- **KITE API URL**: For testing log parsing only, the KITE API URL does not need to be a working endpoint. The service will parse the JSON logs from the file and display results via logs, but webhook sending will fail if the API is not accessible.
- **Log file location**: Ensure the log file path is correct and the file is readable. If `LOG_FILE` is not set, it defaults to `/workspace/shared-data/renovate-logs.json`.
- **Repository format**: The `REPOSITORY` environment variable can use either slashes (e.g., `owner/repo`) or underscores (e.g., `owner_repo`), as underscores are automatically converted to slashes.