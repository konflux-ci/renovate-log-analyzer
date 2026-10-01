// Copyright 2025 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doctor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessLogFile_ExistingTestLogs(t *testing.T) {
	ctx := context.Background()
	result, report, err := ProcessLogFile(ctx, "testdata/test_logs.json")
	if err != nil {
		t.Fatalf("unexpected error processing test_logs.json: %v", err)
	}

	// The test_logs.json contains errors and warnings; verify they were captured
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	// We expect some errors or warnings from the test data
	if len(report.Errors) == 0 && len(report.Warnings) == 0 && result == "" {
		t.Error("expected some findings from test_logs.json but got none")
	}
}

func TestProcessLogFile_FatalExitLogs(t *testing.T) {
	ctx := context.Background()
	result, report, err := ProcessLogFile(ctx, "testdata/fatal_exit_logs.json")
	if err != nil {
		t.Fatalf("unexpected error processing fatal_exit_logs.json: %v", err)
	}

	if report == nil {
		t.Fatal("expected non-nil report")
	}

	// Fatal logs should produce a result string
	if result == "" {
		t.Error("expected non-empty result from fatal_exit_logs.json")
	}
}

func TestProcessLogFile_FileNotFound(t *testing.T) {
	ctx := context.Background()
	_, _, err := ProcessLogFile(ctx, "testdata/nonexistent.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "log file not found") {
		t.Errorf("expected 'log file not found' error, got: %v", err)
	}
}

func TestProcessLogFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.json")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	result, report, err := ProcessLogFile(ctx, emptyFile)
	if err != nil {
		t.Fatalf("unexpected error on empty file: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result for empty file, got: %s", result)
	}
	if len(report.Errors) != 0 || len(report.Warnings) != 0 || len(report.Infos) != 0 {
		t.Error("expected empty report for empty file")
	}
}

func TestProcessLogFile_NormalLines(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "normal.json")

	lines := []string{
		`{"level":30,"msg":"Starting renovate","name":"renovate"}`,
		`{"level":20,"msg":"Processing repo","name":"renovate"}`,
		`{"level":30,"msg":"Finished processing","name":"renovate"}`,
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	result, report, err := ProcessLogFile(ctx, logFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected no errors/fatals, got: %s", result)
	}
	if len(report.Errors) != 0 {
		t.Errorf("expected no report errors, got: %v", report.Errors)
	}
}

func TestProcessLogFile_ErrorAndFatalLines(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "errors.json")

	lines := []string{
		`{"level":30,"msg":"Starting renovate","name":"renovate"}`,
		`{"level":50,"msg":"Something went wrong","name":"renovate"}`,
		`{"level":60,"msg":"Fatal crash","name":"renovate"}`,
		`{"level":30,"msg":"Done","name":"renovate"}`,
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	result, report, err := ProcessLogFile(ctx, logFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result with errors/fatals")
	}
	if !strings.Contains(result, "ERROR") {
		t.Errorf("expected ERROR in result, got: %s", result)
	}
	if !strings.Contains(result, "FATAL") {
		t.Errorf("expected FATAL in result, got: %s", result)
	}
	_ = report // report content depends on selectors; errors/fatals go to result string
}

func TestProcessLogFile_OversizedLine(t *testing.T) {
	// This is the key test: a single JSON line exceeding the 1 MiB buffer.
	// With the old bufio.Scanner, this would fail with "token too long".
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "oversized.json")

	// Build a JSON line that exceeds 1 MiB (1048576 bytes)
	// Create a large "msg" value with padding
	padding := strings.Repeat("x", 2*1024*1024) // 2 MiB of padding
	oversizedLine := fmt.Sprintf(`{"level":50,"msg":"oversized error: %s","name":"renovate"}`, padding)

	// Also include a normal line before and after the oversized one
	lines := []string{
		`{"level":30,"msg":"normal line before","name":"renovate"}`,
		oversizedLine,
		`{"level":30,"msg":"normal line after","name":"renovate"}`,
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	result, _, err := ProcessLogFile(ctx, logFile)
	if err != nil {
		t.Fatalf("ProcessLogFile should handle oversized lines without error, got: %v", err)
	}

	// The oversized line is an ERROR (level 50), so it should appear in the result
	if !strings.Contains(result, "ERROR") {
		t.Errorf("expected ERROR from oversized line in result, got: %s", result)
	}
	if !strings.Contains(result, "oversized error") {
		t.Errorf("expected 'oversized error' message in result, got: %s", result)
	}
}

func TestProcessLogFile_OversizedLineNoNewlineAtEnd(t *testing.T) {
	// Test oversized line at the very end of a file without a trailing newline
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "oversized_no_newline.json")

	padding := strings.Repeat("y", 2*1024*1024)
	oversizedLine := fmt.Sprintf(`{"level":50,"msg":"big trailing error: %s","name":"renovate"}`, padding)

	// No trailing newline — tests the EOF partial-line path
	content := `{"level":30,"msg":"first line","name":"renovate"}` + "\n" + oversizedLine
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	result, _, err := ProcessLogFile(ctx, logFile)
	if err != nil {
		t.Fatalf("ProcessLogFile should handle oversized trailing line without error, got: %v", err)
	}
	if !strings.Contains(result, "big trailing error") {
		t.Errorf("expected 'big trailing error' in result, got: %s", result)
	}
}

func TestProcessLogFile_LastLineWithoutNewline(t *testing.T) {
	// A normal file where the last line has no trailing newline
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "no_trailing_newline.json")

	content := `{"level":50,"msg":"error without newline","name":"renovate"}`
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	result, _, err := ProcessLogFile(ctx, logFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "error without newline") {
		t.Errorf("expected 'error without newline' in result, got: %s", result)
	}
}

func TestProcessLogFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "invalid.json")

	content := "this is not json\n{also bad}\n"
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	result, report, err := ProcessLogFile(ctx, logFile)
	if err != nil {
		t.Fatalf("invalid JSON lines should be skipped, not error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result for invalid JSON, got: %s", result)
	}
	if len(report.Errors) != 0 || len(report.Warnings) != 0 {
		t.Error("expected empty report for invalid JSON")
	}
}

func TestProcessLogFile_Cancellation(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "cancel.json")

	// Write enough lines so the cancellation check fires
	var lines []string
	for i := 0; i < 200; i++ {
		lines = append(lines, `{"level":30,"msg":"line","name":"renovate"}`)
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, _, err := ProcessLogFile(ctx, logFile)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected 'cancelled' in error, got: %v", err)
	}
}

func TestProcessLogFile_SelectorMatch(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "selector.json")

	// Use a message that matches a registered selector
	content := `{"level":40,"msg":"Reached PR limit - skipping PR creation","name":"renovate"}` + "\n"
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, report, err := ProcessLogFile(ctx, logFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Warnings) == 0 {
		t.Error("expected a warning from PR limit selector match")
	}
}

func TestParseLogLine(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantLevel  string
		wantMsg    string
		wantErr    bool
		wantExtras []string
	}{
		{
			name:      "valid error line",
			line:      `{"level":50,"msg":"test error","name":"renovate"}`,
			wantLevel: "ERROR",
			wantMsg:   "test error",
		},
		{
			name:      "valid info line",
			line:      `{"level":30,"msg":"info message","name":"renovate"}`,
			wantLevel: "INFO",
			wantMsg:   "info message",
		},
		{
			name:      "valid fatal line",
			line:      `{"level":60,"msg":"fatal","name":"renovate"}`,
			wantLevel: "FATAL",
			wantMsg:   "fatal",
		},
		{
			name:       "line with extras",
			line:       `{"level":50,"msg":"err msg","branch":"main","depName":"foo"}`,
			wantLevel:  "ERROR",
			wantMsg:    "err msg",
			wantExtras: []string{"branch", "depName"},
		},
		{
			name:    "invalid JSON",
			line:    `not json`,
			wantErr: true,
		},
		{
			name:      "unknown level",
			line:      `{"level":99,"msg":"unknown"}`,
			wantLevel: "",
			wantMsg:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parseLogLine(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Level != tt.wantLevel {
				t.Errorf("level = %q, want %q", entry.Level, tt.wantLevel)
			}
			if entry.Msg != tt.wantMsg {
				t.Errorf("msg = %q, want %q", entry.Msg, tt.wantMsg)
			}
			for _, key := range tt.wantExtras {
				if _, ok := entry.Extras[key]; !ok {
					t.Errorf("expected extra key %q", key)
				}
			}
		})
	}
}
