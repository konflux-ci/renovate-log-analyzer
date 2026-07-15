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
	"strings"
	"testing"
)

func TestPlatformCommitError(t *testing.T) {
	tests := []struct {
		name         string
		line         *LogEntry
		wantErrors   int
		wantContains string
	}{
		{
			name: "well-formed entry with task and commands",
			line: &LogEntry{
				Msg: "Platform-native commit: unknown error",
				Extras: map[string]any{
					"branch": "main",
					"err": map[string]interface{}{
						"message": "push failed",
						"task": map[string]interface{}{
							"commands": []interface{}{
								"push", "--force", "origin",
							},
						},
					},
				},
			},
			wantErrors:   1,
			wantContains: "push",
		},
		{
			name: "err map has no task key",
			line: &LogEntry{
				Msg: "Platform-native commit: unknown error",
				Extras: map[string]any{
					"branch": "main",
					"err": map[string]interface{}{
						"message": "push failed",
					},
				},
			},
			wantErrors:   1,
			wantContains: "push failed",
		},
		{
			name: "task exists but has no commands",
			line: &LogEntry{
				Msg: "Platform-native commit: unknown error",
				Extras: map[string]any{
					"branch": "main",
					"err": map[string]interface{}{
						"message": "push failed",
						"task": map[string]interface{}{
							"format": "utf-8",
						},
					},
				},
			},
			wantErrors:   1,
			wantContains: "push failed",
		},
		{
			name: "task is not a map type",
			line: &LogEntry{
				Msg: "Platform-native commit: unknown error",
				Extras: map[string]any{
					"branch": "main",
					"err": map[string]interface{}{
						"message": "push failed",
						"task":    "not-a-map",
					},
				},
			},
			wantErrors:   1,
			wantContains: "push failed",
		},
		{
			name: "err is not a map type",
			line: &LogEntry{
				Msg: "Platform-native commit: unknown error",
				Extras: map[string]any{
					"branch": "main",
					"err":    "string-error",
				},
			},
			wantErrors: 0,
		},
		{
			name: "err key is missing",
			line: &LogEntry{
				Msg: "Platform-native commit: unknown error",
				Extras: map[string]any{
					"branch": "main",
				},
			},
			wantErrors: 0,
		},
		{
			name: "task is nil",
			line: &LogEntry{
				Msg: "Platform-native commit: unknown error",
				Extras: map[string]any{
					"branch": "main",
					"err": map[string]interface{}{
						"message": "push failed",
						"task":    nil,
					},
				},
			},
			wantErrors:   1,
			wantContains: "push failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &SimpleReport{}
			// Must not panic
			platformCommitError(tt.line, report)
			if len(report.Errors) != tt.wantErrors {
				t.Errorf("got %d errors, want %d", len(report.Errors), tt.wantErrors)
			}
			if tt.wantContains != "" && len(report.Errors) > 0 {
				if !strings.Contains(report.Errors[0], tt.wantContains) {
					t.Errorf("error %q does not contain %q", report.Errors[0], tt.wantContains)
				}
			}
		})
	}
}
