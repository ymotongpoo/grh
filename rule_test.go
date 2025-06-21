// Copyright 2025 Yoshi Yamaguchi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grh

import (
	"strings"
	"testing"
)

func TestRule_CompilePattern(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{
			name: "simple expected pattern",
			rule: Rule{Expected: "Cookie"},
			wantErr: false,
		},
		{
			name: "explicit pattern",
			rule: Rule{Expected: "jQuery", Pattern: "[jJ][qQ][uU][eE][rR][yY]"},
			wantErr: false,
		},
		{
			name: "multiple patterns",
			rule: Rule{Expected: "ハードウェア", Patterns: []string{"ハードウエアー", "ハードウェアー", "ハードウエア"}},
			wantErr: false,
		},
		{
			name: "regex pattern with slashes",
			rule: Rule{Expected: "（$1）", Pattern: "/\\(([^)]+)\\)/"},
			wantErr: false,
		},
		{
			name: "no pattern or expected",
			rule: Rule{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.CompilePattern()
			if (err != nil) != tt.wantErr {
				t.Errorf("Rule.CompilePattern() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRule_Replace(t *testing.T) {
	tests := []struct {
		name     string
		rule     Rule
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple replacement",
			rule:     Rule{Expected: "Cookie", Pattern: "[Cc]ookie"},
			input:    "This is a cookie",
			expected: "This is a Cookie",
			wantErr:  false,
		},
		{
			name:     "multiple patterns",
			rule:     Rule{Expected: "ハードウェア", Patterns: []string{"ハードウエア", "ハードウエアー"}},
			input:    "ハードウエアの話",
			expected: "ハードウェアの話",
			wantErr:  false,
		},
		{
			name:     "regex with capture group",
			rule:     Rule{Expected: "（$1）", Pattern: "/\\(([^)]+)\\)/"},
			input:    "これは(テスト)です",
			expected: "これは（テスト）です",
			wantErr:  false,
		},
		{
			name:     "no compiled pattern",
			rule:     Rule{Expected: "test"},
			input:    "original text",
			expected: "original text",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rule.Pattern != "" || len(tt.rule.Patterns) > 0 {
				err := tt.rule.CompilePattern()
				if err != nil {
					t.Fatalf("Failed to compile pattern: %v", err)
				}
			}
			
			reader := strings.NewReader(tt.input)
			result, err := tt.rule.Replace(reader)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Rule.Replace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if result != tt.expected {
				t.Errorf("Rule.Replace() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRule_ReplaceString(t *testing.T) {
	tests := []struct {
		name     string
		rule     Rule
		input    string
		expected string
	}{
		{
			name:     "simple replacement",
			rule:     Rule{Expected: "Cookie", Pattern: "[Cc]ookie"},
			input:    "This is a cookie",
			expected: "This is a Cookie",
		},
		{
			name:     "multiple patterns",
			rule:     Rule{Expected: "ハードウェア", Patterns: []string{"ハードウエア", "ハードウエアー"}},
			input:    "ハードウエアの話",
			expected: "ハードウェアの話",
		},
		{
			name:     "regex with capture group",
			rule:     Rule{Expected: "（$1）", Pattern: "/\\(([^)]+)\\)/"},
			input:    "これは(テスト)です",
			expected: "これは（テスト）です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.CompilePattern()
			if err != nil {
				t.Fatalf("Failed to compile pattern: %v", err)
			}
			
			result := tt.rule.ReplaceString(tt.input)
			if result != tt.expected {
				t.Errorf("Rule.ReplaceString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRule_ValidateSpecs(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{
			name: "valid specs",
			rule: Rule{
				Expected: "jQuery",
				Pattern:  "[jJ][qQ][uU][eE][rR][yY]",
				Specs: []Spec{
					{From: "jquery", To: "jQuery"},
					{From: "JQUERY", To: "jQuery"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid specs",
			rule: Rule{
				Expected: "JavaScript",
				Pattern:  "[jJ][aA][vV][aA][sS][cC][rR][iI][pP][tT]",
				Specs: []Spec{
					{From: "JAVASCRIPT", To: "JavaScprit"}, // 期待値が間違っている
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.ValidateSpecs()
			if (err != nil) != tt.wantErr {
				t.Errorf("Rule.ValidateSpecs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
