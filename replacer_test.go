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
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestReplacer_ReplaceString(t *testing.T) {
	// テスト用のConfigを作成
	config := &Config{
		Rules: []Rule{
			{Expected: "Cookie", Pattern: "[Cc]ookie"},
			{Expected: "jQuery", Pattern: "[jJ][qQ][uU][eE][rR][yY]"},
		},
	}

	// ルールをコンパイル
	for i := range config.Rules {
		if err := config.Rules[i].CompilePattern(); err != nil {
			t.Fatalf("Failed to compile rule %d: %v", i, err)
		}
	}

	// テスト用のロガーを作成（出力を抑制）
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // エラーレベル以上のみ出力
	}))

	replacer := NewReplacerWithLogger(config, logger)

	input := "This is a cookie and jquery example"
	result := replacer.ReplaceString(input)

	expected := "This is a Cookie and jQuery example"
	if result.Result != expected {
		t.Errorf("ReplaceString() = %q, want %q", result.Result, expected)
	}

	if !result.Changed {
		t.Error("Expected result.Changed to be true")
	}

	if len(result.Changes) == 0 {
		t.Error("Expected some changes to be recorded")
	}
}

func TestReplacer_ReplaceReader(t *testing.T) {
	config := &Config{
		Rules: []Rule{
			{Expected: "Test", Pattern: "[Tt]est"},
		},
	}

	// ルールをコンパイル
	for i := range config.Rules {
		if err := config.Rules[i].CompilePattern(); err != nil {
			t.Fatalf("Failed to compile rule %d: %v", i, err)
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	replacer := NewReplacerWithLogger(config, logger)

	input := "This is a test"
	reader := strings.NewReader(input)
	result, err := replacer.ReplaceReader(reader)

	if err != nil {
		t.Fatalf("ReplaceReader() error = %v", err)
	}

	expected := "This is a Test"
	if result.Result != expected {
		t.Errorf("ReplaceReader() = %q, want %q", result.Result, expected)
	}
}

func TestReplacer_ReplaceFile(t *testing.T) {
	config := &Config{
		Rules: []Rule{
			{Expected: "Cookie", Pattern: "[Cc]ookie"},
			{Expected: "jQuery", Pattern: "[jJ][qQ][uU][eE][rR][yY]"},
		},
	}

	// ルールをコンパイル
	for i := range config.Rules {
		if err := config.Rules[i].CompilePattern(); err != nil {
			t.Fatalf("Failed to compile rule %d: %v", i, err)
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	replacer := NewReplacerWithLogger(config, logger)

	result, err := replacer.ReplaceFile("testdata/doc/sample.md")
	if err != nil {
		t.Fatalf("ReplaceFile() error = %v", err)
	}

	if !result.Changed {
		t.Error("Expected file content to be changed")
	}

	// 結果に期待される変更が含まれているかチェック
	if !strings.Contains(result.Result, "Cookie") {
		t.Error("Expected result to contain 'Cookie'")
	}

	if !strings.Contains(result.Result, "jQuery") {
		t.Error("Expected result to contain 'jQuery'")
	}
}

func TestReplacer_GenerateDiff(t *testing.T) {
	config := &Config{
		Rules: []Rule{
			{Expected: "Test", Pattern: "[Tt]est"},
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	replacer := NewReplacerWithLogger(config, logger)

	result := &ReplaceResult{
		Original: "This is a test",
		Result:   "This is a Test",
		Changed:  true,
	}

	diff := replacer.GenerateDiff(result, "test.txt")

	if diff == "" {
		t.Error("Expected non-empty diff")
	}

	if !strings.Contains(diff, "---") || !strings.Contains(diff, "+++") {
		t.Error("Expected diff to contain unified diff headers")
	}
}

func TestReplacer_ValidateMarkdown(t *testing.T) {
	config := &Config{}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	replacer := NewReplacerWithLogger(config, logger)

	markdown := `# Title

This is a paragraph.

## Subtitle

- List item 1
- List item 2

` + "```javascript\nconsole.log('test');\n```"

	reader := strings.NewReader(markdown)
	err := replacer.ValidateMarkdown(reader)

	if err != nil {
		t.Errorf("ValidateMarkdown() error = %v", err)
	}
}
