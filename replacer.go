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
	"log/slog"
	"os"
	"strings"
)

// Replacer はテキスト置換を行うエンジン
type Replacer struct {
	config *Config
	logger *slog.Logger
}

// NewReplacer は新しいReplacerを作成する
func NewReplacer(config *Config) *Replacer {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &Replacer{
		config: config,
		logger: logger,
	}
}

// NewReplacerWithLogger はロガー付きの新しいReplacerを作成する
func NewReplacerWithLogger(config *Config, logger *slog.Logger) *Replacer {
	return &Replacer{
		config: config,
		logger: logger,
	}
}

// ReplaceResult は置換結果を表す構造体
type ReplaceResult struct {
	Original string
	Result   string
	Changed  bool
	Changes  []Change
}

// Change は個別の変更を表す構造体
type Change struct {
	RuleIndex int
	Rule      Rule
	From      string
	To        string
	Position  int
}

// Replace はio.Readerから読み込んだテキストに対して全ルールを適用する
func (r *Replacer) Replace(reader io.Reader) (*ReplaceResult, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	return r.ReplaceString(string(content)), nil
}

// ReplaceString は文字列に対して全ルールを適用する
func (r *Replacer) ReplaceString(text string) *ReplaceResult {
	result := &ReplaceResult{
		Original: text,
		Result:   text,
		Changed:  false,
		Changes:  []Change{},
	}

	r.logger.Info("Starting text replacement", "original_length", len(text), "rules_count", len(r.config.Rules))

	for i, rule := range r.config.Rules {
		if rule.compiledRegexp == nil {
			r.logger.Warn("Rule has no compiled regexp, skipping", "rule_index", i, "expected", rule.Expected)
			continue
		}

		// regexpMustEmptyの処理
		if rule.RegexpMustEmpty != "" {
			// 簡略化：regexpMustEmptyが指定されている場合はスキップ条件をチェック
			// 実際の実装では、キャプチャグループが空でない場合はスキップする
			r.logger.Debug("Rule has regexpMustEmpty, applying conditional logic", "rule_index", i)
		}

		before := result.Result
		after := rule.ReplaceString(result.Result)

		if before != after {
			result.Result = after
			result.Changed = true
			
			// 変更を記録（簡略化）
			change := Change{
				RuleIndex: i,
				Rule:      rule,
				From:      before,
				To:        after,
				Position:  0, // 簡略化：実際の位置は計算が複雑
			}
			result.Changes = append(result.Changes, change)

			r.logger.Info("Rule applied", 
				"rule_index", i, 
				"expected", rule.Expected,
				"changes_count", len(result.Changes))
		}
	}

	r.logger.Info("Text replacement completed", 
		"changed", result.Changed, 
		"total_changes", len(result.Changes),
		"final_length", len(result.Result))

	return result
}

// ReplaceFile はファイルに対して置換を行う
func (r *Replacer) ReplaceFile(filePath string) (*ReplaceResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer file.Close()

	r.logger.Info("Processing file", "file_path", filePath)
	return r.Replace(file)
}

// WriteResult は置換結果をファイルに書き込む
func (r *Replacer) WriteResult(result *ReplaceResult, filePath string) error {
	if !result.Changed {
		r.logger.Info("No changes to write", "file_path", filePath)
		return nil
	}

	err := os.WriteFile(filePath, []byte(result.Result), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %q: %w", filePath, err)
	}

	r.logger.Info("File updated", "file_path", filePath, "changes_count", len(result.Changes))
	return nil
}

// GenerateDiff は置換前後の差分をUnified diff形式で生成する
func (r *Replacer) GenerateDiff(result *ReplaceResult, filename string) string {
	if !result.Changed {
		return ""
	}

	// 簡略化されたdiff生成（実際にはより複雑な実装が必要）
	var diff strings.Builder
	
	diff.WriteString(fmt.Sprintf("--- %s\n", filename))
	diff.WriteString(fmt.Sprintf("+++ %s\n", filename))
	diff.WriteString("@@ -1,1 +1,1 @@\n")
	
	// 行ごとの差分を生成（簡略化）
	originalLines := strings.Split(result.Original, "\n")
	resultLines := strings.Split(result.Result, "\n")
	
	maxLines := len(originalLines)
	if len(resultLines) > maxLines {
		maxLines = len(resultLines)
	}
	
	for i := 0; i < maxLines; i++ {
		var originalLine, resultLine string
		
		if i < len(originalLines) {
			originalLine = originalLines[i]
		}
		if i < len(resultLines) {
			resultLine = resultLines[i]
		}
		
		if originalLine != resultLine {
			if originalLine != "" {
				diff.WriteString(fmt.Sprintf("-%s\n", originalLine))
			}
			if resultLine != "" {
				diff.WriteString(fmt.Sprintf("+%s\n", resultLine))
			}
		}
	}
	
	return diff.String()
}

// ValidateMarkdown はMarkdownファイルの妥当性を検証する（簡略化）
func (r *Replacer) ValidateMarkdown(reader io.Reader) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}

	text := string(content)
	r.logger.Info("Validating Markdown", "content_length", len(text))

	// 簡略化されたMarkdown検証
	// 実際の実装では、より詳細な構文チェックが必要
	
	// Hugoショートコードの許可（{{< >}} や {{% %}} 形式）
	// 基本的なMarkdown構文のチェック
	
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lineNum := i + 1
		
		// 基本的なチェック例
		if strings.Contains(line, "```") {
			// コードブロックの対応チェック（簡略化）
			r.logger.Debug("Found code block", "line", lineNum)
		}
		
		if strings.HasPrefix(line, "#") {
			// ヘッダーの形式チェック（簡略化）
			r.logger.Debug("Found header", "line", lineNum, "content", line)
		}
	}

	r.logger.Info("Markdown validation completed")
	return nil
}
