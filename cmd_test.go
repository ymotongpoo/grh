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

func TestCLI_Help(t *testing.T) {
	// grhコマンドをビルド
	cmd := exec.Command("go", "build", "-o", "grh_test", "./cmd/grh")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build grh command: %v", err)
	}
	defer os.Remove("grh_test")

	// ヘルプを表示
	cmd = exec.Command("./grh_test", "-h")
	output, err := cmd.CombinedOutput()
	// ヘルプコマンドは正常終了する
	if err != nil {
		// exit status 0 以外の場合のみエラーとする
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			// ヘルプは通常 exit code 0 で終了するが、flagパッケージは異なる場合がある
		}
	}

	helpText := string(output)
	if !strings.Contains(helpText, "Usage:") {
		t.Error("Help text should contain 'Usage:'")
	}

	if !strings.Contains(helpText, "rules-yaml") {
		t.Error("Help text should contain 'rules-yaml' option")
	}
}

func TestCLI_RulesYAML(t *testing.T) {
	// grhコマンドをビルド
	cmd := exec.Command("go", "build", "-o", "grh_test", "./cmd/grh")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build grh command: %v", err)
	}
	defer os.Remove("grh_test")

	// --rules-yamlオプションをテスト
	cmd = exec.Command("./grh_test", "--rules", "testdata/yaml/simple.yml", "--rules-yaml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v, output: %s", err, output)
	}

	yamlOutput := string(output)
	if !strings.Contains(yamlOutput, "version: 1") {
		t.Error("YAML output should contain 'version: 1'")
	}

	if !strings.Contains(yamlOutput, "rules:") {
		t.Error("YAML output should contain 'rules:'")
	}
}

func TestCLI_RulesJSON(t *testing.T) {
	// grhコマンドをビルド
	cmd := exec.Command("go", "build", "-o", "grh_test", "./cmd/grh")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build grh command: %v", err)
	}
	defer os.Remove("grh_test")

	// --rules-jsonオプションをテスト
	cmd = exec.Command("./grh_test", "--rules", "testdata/yaml/simple.yml", "--rules-json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v, output: %s", err, output)
	}

	jsonOutput := string(output)
	if !strings.Contains(jsonOutput, `"version": 1`) {
		t.Error("JSON output should contain '\"version\": 1'")
	}

	if !strings.Contains(jsonOutput, `"rules":`) {
		t.Error("JSON output should contain '\"rules\":'")
	}
}

func TestCLI_Stdout(t *testing.T) {
	// grhコマンドをビルド
	cmd := exec.Command("go", "build", "-o", "grh_test", "./cmd/grh")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build grh command: %v", err)
	}
	defer os.Remove("grh_test")

	// --stdoutオプションをテスト
	cmd = exec.Command("./grh_test", "--rules", "testdata/yaml/simple.yml", "--stdout", "testdata/doc/sample.md")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v, output: %s", err, output)
	}

	// 標準出力に置換結果が含まれているかチェック（stderrにはログが出力される）
	lines := strings.Split(string(output), "\n")
	var stdoutLines []string
	for _, line := range lines {
		// JSONログ行でない場合は標準出力とみなす
		if !strings.HasPrefix(line, "{") {
			stdoutLines = append(stdoutLines, line)
		}
	}

	stdoutContent := strings.Join(stdoutLines, "\n")
	if !strings.Contains(stdoutContent, "Cookie") {
		t.Error("Stdout should contain replaced text 'Cookie'")
	}
}

func TestCLI_Diff(t *testing.T) {
	// grhコマンドをビルド
	cmd := exec.Command("go", "build", "-o", "grh_test", "./cmd/grh")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build grh command: %v", err)
	}
	defer os.Remove("grh_test")

	// --diffオプションをテスト
	cmd = exec.Command("./grh_test", "--rules", "testdata/yaml/simple.yml", "--diff", "testdata/doc/sample.md")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v, output: %s", err, output)
	}

	diffOutput := string(output)
	// Unified diff形式の基本要素をチェック
	if !strings.Contains(diffOutput, "---") {
		t.Error("Diff output should contain '---'")
	}

	if !strings.Contains(diffOutput, "+++") {
		t.Error("Diff output should contain '+++'")
	}
}

func TestCLI_Replace(t *testing.T) {
	// grhコマンドをビルド
	cmd := exec.Command("go", "build", "-o", "grh_test", "./cmd/grh")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build grh command: %v", err)
	}
	defer os.Remove("grh_test")

	// テスト用の一時ファイルを作成
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")
	originalContent := "This is a cookie and jquery example"
	
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// --replaceオプションをテスト
	cmd = exec.Command("./grh_test", "--rules", "testdata/yaml/simple.yml", "--replace", testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v, output: %s", err, output)
	}

	// ファイルが実際に変更されているかチェック
	modifiedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	modifiedText := string(modifiedContent)
	if !strings.Contains(modifiedText, "Cookie") {
		t.Error("File should contain replaced text 'Cookie'")
	}

	if !strings.Contains(modifiedText, "jQuery") {
		t.Error("File should contain replaced text 'jQuery'")
	}

	if modifiedText == originalContent {
		t.Error("File content should have been modified")
	}
}

func TestCLI_Verify(t *testing.T) {
	// grhコマンドをビルド
	cmd := exec.Command("go", "build", "-o", "grh_test", "./cmd/grh")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build grh command: %v", err)
	}
	defer os.Remove("grh_test")

	// --verifyオプションをテスト
	cmd = exec.Command("./grh_test", "--rules", "testdata/yaml/simple.yml", "--verify", "testdata/doc/sample.md")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v, output: %s", err, output)
	}

	// ログにMarkdown検証の成功メッセージが含まれているかチェック
	logOutput := string(output)
	if !strings.Contains(logOutput, "Markdown validation passed") {
		t.Error("Should contain Markdown validation success message")
	}
}
