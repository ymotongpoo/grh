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
	"regexp"
	"strings"
)

// Config はルールファイル全体の設定を表す構造体
type Config struct {
	Version     int       `yaml:"version"`
	Imports     []Import  `yaml:"imports,omitempty"`
	Rules       []Rule    `yaml:"rules"`
	SourcePaths []string  `yaml:"sourcePaths,omitempty"` // --rules-yaml, --rules-json用
}

// Import は他の設定ファイルのインポート設定を表す構造体
type Import struct {
	Path           string   `yaml:"path,omitempty"`
	DisableImports bool     `yaml:"disableImports,omitempty"`
	IgnoreRules    []string `yaml:"ignoreRules,omitempty"`
}

// Rule は個別の置換ルールを表す構造体
type Rule struct {
	Expected        string `yaml:"expected"`
	Pattern         string `yaml:"pattern,omitempty"`
	Patterns        []string `yaml:"patterns,omitempty"`
	RegexpMustEmpty string `yaml:"regexpMustEmpty,omitempty"`
	Specs           []Spec `yaml:"specs,omitempty"`
	
	// 内部処理用（YAMLには出力されない）
	compiledRegexp *regexp.Regexp `yaml:"-"`
}

// Spec はルールのテストケースを表す構造体
type Spec struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// CompilePattern はルールのパターンを正規表現にコンパイルする
func (r *Rule) CompilePattern() error {
	var pattern string
	
	if r.Pattern != "" {
		pattern = r.Pattern
	} else if len(r.Patterns) > 0 {
		// 複数のパターンがある場合は OR で結合
		escapedPatterns := make([]string, len(r.Patterns))
		for i, p := range r.Patterns {
			escapedPatterns[i] = regexp.QuoteMeta(p)
		}
		pattern = strings.Join(escapedPatterns, "|")
	} else if r.Expected != "" {
		// expectedのみの場合は大文字小文字全角半角の統一パターンを生成
		pattern = r.generateCaseInsensitivePattern()
	} else {
		return fmt.Errorf("no pattern or expected value specified")
	}
	
	// 正規表現の形式を確認（/pattern/ 形式の場合は中身を取り出す）
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		pattern = pattern[1 : len(pattern)-1]
	}
	
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("failed to compile pattern %q: %w", pattern, err)
	}
	
	r.compiledRegexp = compiled
	return nil
}

// generateCaseInsensitivePattern は expected 値から大文字小文字全角半角統一パターンを生成
func (r *Rule) generateCaseInsensitivePattern() string {
	expected := r.Expected
	var pattern strings.Builder
	
	for _, char := range expected {
		switch {
		case (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z'):
			// アルファベットの場合は大文字小文字全角半角のパターンを生成
			upper := strings.ToUpper(string(char))
			lower := strings.ToLower(string(char))
			// 全角文字への変換は簡略化（実際にはより複雑な変換が必要）
			pattern.WriteString(fmt.Sprintf("[%s%s]", upper, lower))
		default:
			// その他の文字はそのまま（エスケープが必要な場合は対応）
			pattern.WriteString(regexp.QuoteMeta(string(char)))
		}
	}
	
	return pattern.String()
}

// ReplaceReader はio.Readerから読み込んだテキストに対してルールを適用して置換を行う
func (r *Rule) ReplaceReader(reader io.Reader) (string, error) {
	if r.compiledRegexp == nil {
		// パターンがコンパイルされていない場合は、そのまま読み込んで返す
		content, err := io.ReadAll(reader)
		if err != nil {
			return "", fmt.Errorf("failed to read from reader: %w", err)
		}
		return string(content), nil
	}
	
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read from reader: %w", err)
	}
	
	return r.compiledRegexp.ReplaceAllString(string(content), r.Expected), nil
}

// Replace はテキストに対してルールを適用して置換を行う（後方互換性のため残す）
func (r *Rule) Replace(text string) string {
	if r.compiledRegexp == nil {
		return text
	}
	
	return r.compiledRegexp.ReplaceAllString(text, r.Expected)
}

// ValidateSpecs はルールのテストケースを検証する
func (r *Rule) ValidateSpecs() error {
	if r.compiledRegexp == nil {
		if err := r.CompilePattern(); err != nil {
			return err
		}
	}
	
	for _, spec := range r.Specs {
		result := r.Replace(spec.From)
		if result != spec.To {
			return fmt.Errorf("spec failed: %q expected %q, but got %q", spec.From, spec.To, result)
		}
	}
	
	return nil
}
