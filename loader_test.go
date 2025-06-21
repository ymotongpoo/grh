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
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("testdata/yaml/simple.yml")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Version = %d, want 1", config.Version)
	}

	if len(config.Rules) != 3 {
		t.Errorf("len(Rules) = %d, want 3", len(config.Rules))
	}

	// 最初のルールをテスト
	rule := config.Rules[0]
	if rule.Expected != "Cookie" {
		t.Errorf("Rules[0].Expected = %q, want %q", rule.Expected, "Cookie")
	}

	// ルールが正しく動作するかテスト
	result := rule.ReplaceString("This is a cookie")
	expected := "This is a Cookie"
	if result != expected {
		t.Errorf("Rules[0].ReplaceString() = %q, want %q", result, expected)
	}
}

func TestLoadConfigFromReader(t *testing.T) {
	yamlContent := `version: 1
rules:
  - expected: Test
    specs:
      - from: test
        to: Test`

	reader := strings.NewReader(yamlContent)
	config, err := LoadConfigFromReader(reader, "test.yml")
	if err != nil {
		t.Fatalf("LoadConfigFromReader() error = %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Version = %d, want 1", config.Version)
	}

	if len(config.Rules) != 1 {
		t.Errorf("len(Rules) = %d, want 1", len(config.Rules))
	}

	if len(config.SourcePaths) != 1 || config.SourcePaths[0] != "test.yml" {
		t.Errorf("SourcePaths = %v, want [test.yml]", config.SourcePaths)
	}
}

func TestFindRuleFile(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	
	// prh.ymlファイルを作成
	prhPath := filepath.Join(tempDir, "prh.yml")
	err := os.WriteFile(prhPath, []byte("version: 1\nrules: []"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// サブディレクトリを作成
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// サブディレクトリから検索
	found, err := FindRuleFile(subDir)
	if err != nil {
		t.Fatalf("FindRuleFile() error = %v", err)
	}

	if found != prhPath {
		t.Errorf("FindRuleFile() = %q, want %q", found, prhPath)
	}
}

func TestMergeConfigs(t *testing.T) {
	config1 := &Config{
		Version: 1,
		Rules: []Rule{
			{Expected: "Rule1"},
		},
		SourcePaths: []string{"config1.yml"},
	}

	config2 := &Config{
		Version: 1,
		Rules: []Rule{
			{Expected: "Rule2"},
			{Expected: "Rule1"}, // 同じexpectedで上書き
		},
		SourcePaths: []string{"config2.yml"},
	}

	merged := MergeConfigs(config1, config2)

	if len(merged.SourcePaths) != 2 {
		t.Errorf("len(SourcePaths) = %d, want 2", len(merged.SourcePaths))
	}

	// ルールの数は重複を除いて2つになるはず
	if len(merged.Rules) != 2 {
		t.Errorf("len(Rules) = %d, want 2", len(merged.Rules))
	}
}
