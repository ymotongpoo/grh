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
	Version     int       `yaml:"version" json:"version"`
	Imports     []Import  `yaml:"imports,omitempty" json:"imports,omitempty"`
	Rules       []Rule    `yaml:"rules" json:"rules"`
	SourcePaths []string  `yaml:"sourcePaths,omitempty" json:"sourcePaths,omitempty"` // --rules-yaml, --rules-json用
}

// Import は他の設定ファイルのインポート設定を表す構造体
type Import struct {
	Path           string   `yaml:"path,omitempty" json:"path,omitempty"`
	DisableImports bool     `yaml:"disableImports,omitempty" json:"disableImports,omitempty"`
	IgnoreRules    []string `yaml:"ignoreRules,omitempty" json:"ignoreRules,omitempty"`
}

// Rule は個別の置換ルールを表す構造体
type Rule struct {
	Expected            string   `yaml:"expected" json:"expected"`
	Pattern             string   `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Patterns            []string `yaml:"patterns,omitempty" json:"patterns,omitempty"`
	RegexpMustEmpty     string   `yaml:"regexpMustEmpty,omitempty" json:"regexpMustEmpty,omitempty"`
	Specs               []Spec   `yaml:"specs,omitempty" json:"specs,omitempty"`
	IgnorePatternBefore string   `yaml:"ignorePatternBefore,omitempty" json:"ignorePatternBefore,omitempty"`

	// 内部処理用（YAMLには出力されない）
	compiledRegexp       *regexp.Regexp `yaml:"-" json:"-"`
	compiledIgnoreBefore *regexp.Regexp `yaml:"-" json:"-"`
}

// Spec はルールのテストケースを表す構造体
type Spec struct {
	From string `yaml:"from" json:"from"`
	To   string `yaml:"to" json:"to"`
}

// CompilePattern はルールのパターンを正規表現にコンパイルする
func (r *Rule) CompilePattern() error {
	var pattern string
	
	if r.Pattern != "" {
		pattern = r.Pattern
	} else if len(r.Patterns) > 0 {
		// 複数のパターンがある場合は OR で結合
		// patternsは正規表現として扱う（エスケープしない）
		pattern = strings.Join(r.Patterns, "|")
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

	if r.IgnorePatternBefore != "" {
		ignorePattern := r.IgnorePatternBefore
		
		// パターンが既に$で終わっている場合や、$を含む複雑なパターンの場合は追加しない
		// また、( |$)のような選択パターンを含む場合も追加しない
		if !strings.HasSuffix(ignorePattern, "$") && !strings.Contains(ignorePattern, "$") && !strings.Contains(ignorePattern, "|") {
			ignorePattern += "$"
		}
		
		compiledIgnore, err := regexp.Compile(ignorePattern)
		if err != nil {
			return fmt.Errorf("failed to compile ignorePatternBefore %q: %w", r.IgnorePatternBefore, err)
		}
		r.compiledIgnoreBefore = compiledIgnore
	}
	return nil
}

// generateCaseInsensitivePattern は expected 値から大文字小文字全角半角統一パターンを生成
func (r *Rule) generateCaseInsensitivePattern() string {
	expected := r.Expected
	var pattern strings.Builder
	
	for _, char := range expected {
		switch {
		case (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z'):
			// 半角アルファベットの場合は大文字小文字全角半角のパターンを生成
			upper := strings.ToUpper(string(char))
			lower := strings.ToLower(string(char))
			
			// 全角文字への変換
			fullwidthUpper := r.toFullwidthAlphabet(upper)
			fullwidthLower := r.toFullwidthAlphabet(lower)
			
			pattern.WriteString(fmt.Sprintf("[%s%s%s%s]", upper, lower, fullwidthUpper, fullwidthLower))
		case (char >= 'Ａ' && char <= 'Ｚ') || (char >= 'ａ' && char <= 'ｚ'):
			// 全角アルファベットの場合は大文字小文字半角全角のパターンを生成
			halfwidthChar := r.toHalfwidthAlphabet(string(char))
			upper := strings.ToUpper(halfwidthChar)
			lower := strings.ToLower(halfwidthChar)
			fullwidthUpper := r.toFullwidthAlphabet(upper)
			fullwidthLower := r.toFullwidthAlphabet(lower)
			
			pattern.WriteString(fmt.Sprintf("[%s%s%s%s]", upper, lower, fullwidthUpper, fullwidthLower))
		default:
			// その他の文字はそのまま（エスケープが必要な場合は対応）
			pattern.WriteString(regexp.QuoteMeta(string(char)))
		}
	}
	
	return pattern.String()
}

// toFullwidthAlphabet は半角アルファベットを全角アルファベットに変換
func (r *Rule) toFullwidthAlphabet(s string) string {
	var result strings.Builder
	for _, char := range s {
		if char >= 'A' && char <= 'Z' {
			// A-Z を Ａ-Ｚ に変換
			result.WriteRune('Ａ' + (char - 'A'))
		} else if char >= 'a' && char <= 'z' {
			// a-z を ａ-ｚ に変換
			result.WriteRune('ａ' + (char - 'a'))
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// toHalfwidthAlphabet は全角アルファベットを半角アルファベットに変換
func (r *Rule) toHalfwidthAlphabet(s string) string {
	var result strings.Builder
	for _, char := range s {
		if char >= 'Ａ' && char <= 'Ｚ' {
			// Ａ-Ｚ を A-Z に変換
			result.WriteRune('A' + (char - 'Ａ'))
		} else if char >= 'ａ' && char <= 'ｚ' {
			// ａ-ｚ を a-z に変換
			result.WriteRune('a' + (char - 'ａ'))
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// Replace はio.Readerから読み込んだテキストに対してルールを適用して置換を行う
func (r *Rule) Replace(reader io.Reader) (string, error) {
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

// ReplaceString はテキストに対してルールを適用して置換を行う
func (r *Rule) ReplaceString(text string) string {
	if r.compiledRegexp == nil {
		return text
	}
	
	// ignorePatternBeforeが設定されている場合の処理
	if r.compiledIgnoreBefore != nil {
		var sb strings.Builder
		lastIndex := 0
		matches := r.compiledRegexp.FindAllStringIndex(text, -1)

		for _, matchIndices := range matches {
			startIndex := matchIndices[0]
			endIndex := matchIndices[1]

			sb.WriteString(text[lastIndex:startIndex])

			contextBefore := text[:startIndex]

			shouldIgnore := r.compiledIgnoreBefore.MatchString(contextBefore)

			if shouldIgnore {
				sb.WriteString(text[startIndex:endIndex])
			} else {
				sb.WriteString(r.Expected)
			}
			lastIndex = endIndex
		}
		sb.WriteString(text[lastIndex:])
		result := sb.String()
		
		return result
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
		result := r.ReplaceString(spec.From)
		if result != spec.To {
			return fmt.Errorf("spec failed: %q expected %q, but got %q", spec.From, spec.To, result)
		}
	}
	
	return nil
}
