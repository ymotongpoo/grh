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

func TestIntegration_ComplexRules(t *testing.T) {
	config, err := LoadConfig("testdata/yaml/complex.yml")
	if err != nil {
		t.Fatalf("Failed to load complex rules: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	replacer := NewReplacerWithLogger(config, logger)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JavaScript conversion",
			input:    "I love javascript programming",
			expected: "I love JavaScript programming",
		},
		{
			name:     "Database conversion",
			input:    "データーベースとDBの設計",
			expected: "データベースとデータベースの設計",
		},
		{
			name:     "Computer conversion",
			input:    "コンピュータの性能",
			expected: "コンピューターの性能",
		},
		{
			name:     "API conversion",
			input:    "REST apiの設計",
			expected: "REST APIの設計",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacer.ReplaceString(tt.input)
			if result.Result != tt.expected {
				t.Errorf("ReplaceString() = %q, want %q", result.Result, tt.expected)
			}
		})
	}
}

func TestIntegration_WithImports(t *testing.T) {
	// testdata/yaml/with-imports.ymlはbase.ymlをインポートする
	config, err := LoadConfigWithImports("testdata/yaml/with-imports.yml")
	if err != nil {
		t.Fatalf("Failed to load config with imports: %v", err)
	}

	// インポートされたルールとローカルルールの両方が含まれているかチェック
	expectedRuleCount := 4 // base.yml(2) + with-imports.yml(2)
	if len(config.Rules) != expectedRuleCount {
		t.Errorf("Expected %d rules, got %d", expectedRuleCount, len(config.Rules))
	}

	// ソースパスが両方記録されているかチェック
	if len(config.SourcePaths) < 2 {
		t.Errorf("Expected at least 2 source paths, got %d", len(config.SourcePaths))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	replacer := NewReplacerWithLogger(config, logger)

	// インポートされたルールのテスト
	result := replacer.ReplaceString("html and css")
	expected := "HTML and CSS"
	if result.Result != expected {
		t.Errorf("Imported rules failed: got %q, want %q", result.Result, expected)
	}

	// ローカルルールのテスト
	result = replacer.ReplaceString("react and vue")
	expected = "React and Vue.js"
	if result.Result != expected {
		t.Errorf("Local rules failed: got %q, want %q", result.Result, expected)
	}
}

func TestIntegration_FileProcessing(t *testing.T) {
	config, err := LoadConfig("testdata/yaml/complex.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	replacer := NewReplacerWithLogger(config, logger)

	result, err := replacer.ReplaceFile("testdata/doc/complex.md")
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}

	if !result.Changed {
		t.Error("Expected file to be changed")
	}

	// 期待される変更がすべて含まれているかチェック
	expectedChanges := []string{
		"JavaScript", // javascript -> JavaScript
		"データベース",   // データーベース -> データベース
		"API",        // api -> API
		"コンピューター",   // コンピュータ -> コンピューター
	}

	for _, expected := range expectedChanges {
		if !strings.Contains(result.Result, expected) {
			t.Errorf("Expected result to contain %q", expected)
		}
	}

	// Hugoショートコードが保持されているかチェック
	if !strings.Contains(result.Result, "{{< highlight") {
		t.Error("Hugo shortcode should be preserved")
	}

	if !strings.Contains(result.Result, "{{% note %}") {
		t.Error("Hugo shortcode should be preserved")
	}
}

func TestIntegration_MarkdownValidation(t *testing.T) {
	config := &Config{} // 空の設定でMarkdown検証のみテスト

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	replacer := NewReplacerWithLogger(config, logger)

	file, err := os.Open("testdata/doc/complex.md")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	err = replacer.ValidateMarkdown(file)
	if err != nil {
		t.Errorf("Markdown validation failed: %v", err)
	}
}

func TestIntegration_DiffGeneration(t *testing.T) {
	config, err := LoadConfig("testdata/yaml/simple.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	replacer := NewReplacerWithLogger(config, logger)

	original := "This is a cookie and jquery example"
	result := replacer.ReplaceString(original)

	diff := replacer.GenerateDiff(result, "test.txt")

	if diff == "" {
		t.Error("Expected non-empty diff")
	}

	// Unified diff形式の基本的な要素をチェック
	if !strings.Contains(diff, "---") {
		t.Error("Diff should contain '---' header")
	}

	if !strings.Contains(diff, "+++") {
		t.Error("Diff should contain '+++' header")
	}

	if !strings.Contains(diff, "Cookie") {
		t.Error("Diff should show the change to 'Cookie'")
	}

	if !strings.Contains(diff, "jQuery") {
		t.Error("Diff should show the change to 'jQuery'")
	}
}
