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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFinalIntegration_FullWorkflow は全体的なワークフローをテストする
func TestFinalIntegration_FullWorkflow(t *testing.T) {
	// grhコマンドをビルド
	cmd := exec.Command("go", "build", "-o", "grh_integration_test", "./cmd/grh")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build grh command: %v", err)
	}
	defer os.Remove("grh_integration_test")

	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// テスト用のルールファイルを作成
	ruleFile := filepath.Join(tempDir, "prh.yml")
	ruleContent := `version: 1
rules:
  - expected: JavaScript
    pattern: "[jJ][aA][vV][aA][sS][cC][rR][iI][pP][tT]"
    specs:
      - from: javascript
        to: JavaScript
  - expected: Cookie
    specs:
      - from: cookie
        to: Cookie
  - expected: データベース
    patterns:
      - データーベース
      - DB
    specs:
      - from: データーベース
        to: データベース
`
	err := os.WriteFile(ruleFile, []byte(ruleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create rule file: %v", err)
	}

	// テスト用のMarkdownファイルを作成
	testFile := filepath.Join(tempDir, "test.md")
	testContent := `# Web開発ガイド

このドキュメントはjavascriptとcookieを使ったweb開発について説明します。

## 通常のテキスト

通常のテキストではjavascriptとcookieが置換されます。
データーベースの設計も重要です。

## コードブロック

` + "```javascript" + `
// このコードブロック内のjavascriptとcookieは置換される
console.log('javascript cookie example');
` + "```" + `
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 1. --rules-yamlオプションのテスト
	t.Run("rules-yaml output", func(t *testing.T) {
		cmd := exec.Command("./grh_integration_test", "--rules", ruleFile, "--rules-yaml")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v, output: %s", err, output)
		}

		yamlOutput := string(output)
		if !strings.Contains(yamlOutput, "version: 1") {
			t.Error("YAML output should contain version")
		}
		if !strings.Contains(yamlOutput, "expected: JavaScript") {
			t.Error("YAML output should contain JavaScript rule")
		}
	})

	// 2. --stdoutオプションのテスト
	t.Run("stdout processing", func(t *testing.T) {
		cmd := exec.Command("./grh_integration_test", "--rules", ruleFile, "--stdout", testFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v, output: %s", err, output)
		}

		// ログを除いた標準出力を取得
		lines := strings.Split(string(output), "\n")
		var stdoutLines []string
		for _, line := range lines {
			if !strings.HasPrefix(line, "{") { // JSONログ行でない
				stdoutLines = append(stdoutLines, line)
			}
		}
		result := strings.Join(stdoutLines, "\n")

		// 通常のテキストが置換されているかチェック
		if !strings.Contains(result, "JavaScript") {
			t.Error("Should contain replaced 'JavaScript'")
		}
		if !strings.Contains(result, "Cookie") {
			t.Error("Should contain replaced 'Cookie'")
		}
		if !strings.Contains(result, "データベース") {
			t.Error("Should contain replaced 'データベース'")
		}
	})

	// 3. --diffオプションのテスト
	t.Run("diff output", func(t *testing.T) {
		cmd := exec.Command("./grh_integration_test", "--rules", ruleFile, "--diff", testFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v, output: %s", err, output)
		}

		diffOutput := string(output)
		if !strings.Contains(diffOutput, "---") || !strings.Contains(diffOutput, "+++") {
			t.Error("Diff output should contain unified diff headers")
		}
	})

	// 4. --replaceオプションのテスト
	t.Run("file replacement", func(t *testing.T) {
		// テストファイルのコピーを作成
		copyFile := filepath.Join(tempDir, "test_copy.md")
		originalContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read original file: %v", err)
		}
		err = os.WriteFile(copyFile, originalContent, 0644)
		if err != nil {
			t.Fatalf("Failed to create copy file: %v", err)
		}

		// --replaceオプションを実行
		cmd := exec.Command("./grh_integration_test", "--rules", ruleFile, "--replace", copyFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v, output: %s", err, output)
		}

		// ファイルが実際に変更されているかチェック
		modifiedContent, err := os.ReadFile(copyFile)
		if err != nil {
			t.Fatalf("Failed to read modified file: %v", err)
		}

		modifiedText := string(modifiedContent)
		if !strings.Contains(modifiedText, "JavaScript") {
			t.Error("File should contain replaced 'JavaScript'")
		}
		if !strings.Contains(modifiedText, "Cookie") {
			t.Error("File should contain replaced 'Cookie'")
		}

		// 元のファイルと異なることを確認
		if string(originalContent) == modifiedText {
			t.Error("File content should have been modified")
		}
	})

	// 5. --verifyオプションのテスト
	t.Run("markdown verification", func(t *testing.T) {
		cmd := exec.Command("./grh_integration_test", "--rules", ruleFile, "--verify", testFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v, output: %s", err, output)
		}

		logOutput := string(output)
		if !strings.Contains(logOutput, "Markdown validation completed") {
			t.Error("Should contain validation completion message")
		}
	})
}

// TestFinalIntegration_RuleFileDiscovery はルールファイル自動検索をテストする
func TestFinalIntegration_RuleFileDiscovery(t *testing.T) {
	// grhコマンドをビルド
	cmd := exec.Command("go", "build", "-o", "grh_integration_test", "./cmd/grh")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build grh command: %v", err)
	}
	defer os.Remove("grh_integration_test")

	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// prh.ymlファイルを作成
	ruleFile := filepath.Join(tempDir, "prh.yml")
	ruleContent := `version: 1
rules:
  - expected: Test
    specs:
      - from: test
        to: Test
`
	err := os.WriteFile(ruleFile, []byte(ruleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create rule file: %v", err)
	}

	// テストファイルを作成
	testFile := filepath.Join(tempDir, "sample.md")
	testContent := "This is a test document."
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 作業ディレクトリを変更してコマンドを実行
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// ルールファイルを指定せずにコマンドを実行（自動検索）
	cmd = exec.Command(filepath.Join(originalDir, "grh_integration_test"), "--stdout", "sample.md")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v, output: %s", err, output)
	}

	// 置換が実行されているかチェック
	if !strings.Contains(string(output), "Test") {
		t.Error("Rule file should be automatically discovered and applied")
	}
}
func extractBetween(text, start, end string) string {
	startIdx := strings.Index(text, start)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(start)

	endIdx := strings.Index(text[startIdx:], end)
	if endIdx == -1 {
		return ""
	}

	return text[startIdx : startIdx+endIdx]
}
