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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfig はYAMLファイルからConfigを読み込む
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", path, err)
	}
	defer file.Close()

	return LoadConfigFromReader(file, path)
}

// LoadConfigFromReader はio.ReaderからConfigを読み込む
func LoadConfigFromReader(reader io.Reader, sourcePath string) (*Config, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// ソースパスを記録
	config.SourcePaths = []string{sourcePath}

	// ルールのパターンをコンパイル
	for i := range config.Rules {
		if err := config.Rules[i].CompilePattern(); err != nil {
			return nil, fmt.Errorf("failed to compile pattern for rule %d: %w", i, err)
		}
	}

	// ルールのテストケースを検証
	for i, rule := range config.Rules {
		if err := rule.ValidateSpecs(); err != nil {
			return nil, fmt.Errorf("rule %d validation failed: %w", i, err)
		}
	}

	return &config, nil
}

// FindRuleFile はカレントディレクトリから上位ディレクトリに向かってprh.yml/prh.yamlを探す
func FindRuleFile(startDir string) (string, error) {
	dir := startDir
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	for {
		// prh.yml を優先的に探す
		for _, filename := range []string{"prh.yml", "prh.yaml"} {
			path := filepath.Join(dir, filename)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}

		// 親ディレクトリに移動
		parent := filepath.Dir(dir)
		if parent == dir {
			// ルートディレクトリに到達した
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("rule file (prh.yml or prh.yaml) not found")
}

// MergeConfigs は複数のConfigをマージする（後のものが優先される）
func MergeConfigs(configs ...*Config) *Config {
	if len(configs) == 0 {
		return &Config{}
	}

	merged := &Config{
		Version: configs[0].Version,
		Rules:   []Rule{},
	}

	// すべてのソースパスを収集
	var sourcePaths []string
	for _, config := range configs {
		sourcePaths = append(sourcePaths, config.SourcePaths...)
	}
	merged.SourcePaths = sourcePaths

	// ルールをマージ（後のものが優先）
	ruleMap := make(map[string]Rule)
	for _, config := range configs {
		for _, rule := range config.Rules {
			// ルールのキーとしてexpectedを使用（簡略化）
			key := rule.Expected
			ruleMap[key] = rule
		}
	}

	// マップからスライスに変換
	for _, rule := range ruleMap {
		merged.Rules = append(merged.Rules, rule)
	}

	return merged
}

// LoadConfigWithImports はインポートを含むConfigを読み込む
func LoadConfigWithImports(path string) (*Config, error) {
	config, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}

	if len(config.Imports) == 0 {
		return config, nil
	}

	configs := []*Config{config}
	baseDir := filepath.Dir(path)

	for _, imp := range config.Imports {
		var importPath string
		if filepath.IsAbs(imp.Path) {
			importPath = imp.Path
		} else {
			importPath = filepath.Join(baseDir, imp.Path)
		}

		var importedConfig *Config
		if imp.DisableImports {
			// インポートの連鎖を無効にする場合は直接読み込み
			importedConfig, err = LoadConfig(importPath)
		} else {
			// 再帰的にインポートを処理
			importedConfig, err = LoadConfigWithImports(importPath)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to load imported config %q: %w", importPath, err)
		}

		// ignoreRulesの処理（簡略化：expectedでマッチング）
		if len(imp.IgnoreRules) > 0 {
			filteredRules := []Rule{}
			for _, rule := range importedConfig.Rules {
				ignore := false
				for _, ignorePattern := range imp.IgnoreRules {
					if strings.Contains(rule.Expected, ignorePattern) {
						ignore = true
						break
					}
				}
				if !ignore {
					filteredRules = append(filteredRules, rule)
				}
			}
			importedConfig.Rules = filteredRules
		}

		configs = append(configs, importedConfig)
	}

	return MergeConfigs(configs...), nil
}
