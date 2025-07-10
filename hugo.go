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
	"regexp"
	"strings"
)

// HugoShortcode はHugoショートコードの情報を表す構造体
type HugoShortcode struct {
	Type     string // "paired" または "self-closing"
	Name     string
	Content  string
	Position int
	Length   int
}

// HugoProcessor はHugoショートコードとMarkdownコードの処理を行う
type HugoProcessor struct {
	// Hugoショートコードのパターン
	pairedShortcodeRegex      *regexp.Regexp // {{< name >}}...{{< /name >}}
	pairedShortcodeRegexAlt   *regexp.Regexp // {{% name %}}...{{% /name %}}
	selfClosingShortcodeRegex *regexp.Regexp // {{< name />}} または {{< name >}}
	selfClosingShortcodeRegexAlt *regexp.Regexp // {{% name /%}} または {{% name %}}
	
	// Markdownコードのパターン
	codeBlockRegex *regexp.Regexp // ```...```
	codeSpanRegex  *regexp.Regexp // `...`
	
	// Markdownリンクのパターン
	linkRegex      *regexp.Regexp // [text](url)
	refLinkRegex   *regexp.Regexp // [ref]: url
	refLinkUseRegex *regexp.Regexp // [text][ref]
}

// NewHugoProcessor は新しいHugoProcessorを作成する
func NewHugoProcessor() *HugoProcessor {
	return &HugoProcessor{
		// ペアードショートコード: {{< name >}}...{{< /name >}}
		// 注意：RE2では後方参照が使えないため、より単純なパターンを使用
		pairedShortcodeRegex: regexp.MustCompile(`(?s)\{\{<\s*([a-zA-Z0-9_-]+)(?:\s+[^>]*)?\s*>\}\}(.*?)\{\{<\s*/[a-zA-Z0-9_-]+\s*>\}\}`),
		// ペアードショートコード（代替形式）: {{% name %}}...{{% /name %}}
		pairedShortcodeRegexAlt: regexp.MustCompile(`(?s)\{\{%\s*([a-zA-Z0-9_-]+)(?:\s+[^%]*)?\s*%\}\}(.*?)\{\{%\s*/[a-zA-Z0-9_-]+\s*%\}\}`),
		// セルフクローズショートコード: {{< name />}} または {{< name >}}
		selfClosingShortcodeRegex: regexp.MustCompile(`\{\{<\s*([a-zA-Z0-9_-]+)(?:\s+[^>]*)?\s*/?>\}\}`),
		// セルフクローズショートコード（代替形式）: {{% name /%}} または {{% name %}}
		selfClosingShortcodeRegexAlt: regexp.MustCompile(`\{\{%\s*([a-zA-Z0-9_-]+)(?:\s+[^%]*)?\s*/?%\}\}`),
		
		// Markdownコードブロック: ```...```（複数行対応）
		codeBlockRegex: regexp.MustCompile("(?s)```[^\\n]*\\n(.*?)```"),
		// Markdownコードスパン: `...`（単一行）
		codeSpanRegex: regexp.MustCompile("`([^`\n]+)`"),
		
		// Markdownインラインリンク: [text](url)
		linkRegex: regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`),
		// Markdown参照リンク定義: [ref]: url
		refLinkRegex: regexp.MustCompile(`(?m)^\s*\[([^\]]+)\]:\s*(.+)$`),
		// Markdown参照リンク使用: [text][ref]
		refLinkUseRegex: regexp.MustCompile(`\[([^\]]*)\]\[([^\]]+)\]`),
	}
}

// FindShortcodes はテキスト内のHugoショートコードを検出する
func (hp *HugoProcessor) FindShortcodes(text string) []HugoShortcode {
	var shortcodes []HugoShortcode

	// ペアードショートコード（{{< >}}形式）を検索
	matches := hp.pairedShortcodeRegex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		if len(match) >= 6 {
			shortcode := HugoShortcode{
				Type:     "paired",
				Name:     text[match[2]:match[3]],
				Content:  text[match[4]:match[5]],
				Position: match[0],
				Length:   match[1] - match[0],
			}
			shortcodes = append(shortcodes, shortcode)
		}
	}

	// ペアードショートコード（{{% %}}形式）を検索
	matches = hp.pairedShortcodeRegexAlt.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		if len(match) >= 6 {
			shortcode := HugoShortcode{
				Type:     "paired",
				Name:     text[match[2]:match[3]],
				Content:  text[match[4]:match[5]],
				Position: match[0],
				Length:   match[1] - match[0],
			}
			shortcodes = append(shortcodes, shortcode)
		}
	}

	// セルフクローズショートコード（{{< >}}形式）を検索
	matches = hp.selfClosingShortcodeRegex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			// ペアードショートコードと重複しないかチェック
			if !hp.isPartOfPairedShortcode(text, match[0], shortcodes) {
				shortcode := HugoShortcode{
					Type:     "self-closing",
					Name:     text[match[2]:match[3]],
					Content:  "",
					Position: match[0],
					Length:   match[1] - match[0],
				}
				shortcodes = append(shortcodes, shortcode)
			}
		}
	}

	// セルフクローズショートコード（{{% %}}形式）を検索
	matches = hp.selfClosingShortcodeRegexAlt.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			// ペアードショートコードと重複しないかチェック
			if !hp.isPartOfPairedShortcode(text, match[0], shortcodes) {
				shortcode := HugoShortcode{
					Type:     "self-closing",
					Name:     text[match[2]:match[3]],
					Content:  "",
					Position: match[0],
					Length:   match[1] - match[0],
				}
				shortcodes = append(shortcodes, shortcode)
			}
		}
	}

	return shortcodes
}

// isPartOfPairedShortcode は指定位置がペアードショートコードの一部かどうかをチェック
func (hp *HugoProcessor) isPartOfPairedShortcode(text string, position int, shortcodes []HugoShortcode) bool {
	for _, sc := range shortcodes {
		if sc.Type == "paired" && position >= sc.Position && position < sc.Position+sc.Length {
			return true
		}
	}
	return false
}

// PreserveShortcodes はショートコードとMarkdownコードを一時的にプレースホルダーに置換する
func (hp *HugoProcessor) PreserveShortcodes(text string) (string, map[string]string) {
	placeholders := make(map[string]string)
	result := text

	// より安全なアプローチ：正規表現で直接置換
	counter := 0
	
	// Markdownコードブロックを最初に保護（優先度高）
	result = hp.codeBlockRegex.ReplaceAllStringFunc(result, func(match string) string {
		counter++
		placeholder := fmt.Sprintf("___MARKDOWN_CODE_BLOCK_%d___", counter)
		placeholders[placeholder] = match
		return placeholder
	})
	
	// Markdownコードスパンを保護
	result = hp.codeSpanRegex.ReplaceAllStringFunc(result, func(match string) string {
		counter++
		placeholder := fmt.Sprintf("___MARKDOWN_CODE_SPAN_%d___", counter)
		placeholders[placeholder] = match
		return placeholder
	})
	
	// Markdownインラインリンクを保護
	result = hp.linkRegex.ReplaceAllStringFunc(result, func(match string) string {
		counter++
		placeholder := fmt.Sprintf("___MARKDOWN_LINK_%d___", counter)
		placeholders[placeholder] = match
		return placeholder
	})
	
	// Markdown参照リンク使用を保護
	result = hp.refLinkUseRegex.ReplaceAllStringFunc(result, func(match string) string {
		counter++
		placeholder := fmt.Sprintf("___MARKDOWN_REF_LINK_USE_%d___", counter)
		placeholders[placeholder] = match
		return placeholder
	})
	
	// Markdown参照リンク定義を保護
	result = hp.refLinkRegex.ReplaceAllStringFunc(result, func(match string) string {
		counter++
		placeholder := fmt.Sprintf("___MARKDOWN_REF_LINK_%d___", counter)
		placeholders[placeholder] = match
		return placeholder
	})
	
	// ペアードショートコード（{{< >}}形式）を置換
	result = hp.pairedShortcodeRegex.ReplaceAllStringFunc(result, func(match string) string {
		counter++
		placeholder := fmt.Sprintf("___HUGO_SHORTCODE_PAIRED_%d___", counter)
		placeholders[placeholder] = match
		return placeholder
	})

	// ペアードショートコード（{{% %}}形式）を置換
	result = hp.pairedShortcodeRegexAlt.ReplaceAllStringFunc(result, func(match string) string {
		counter++
		placeholder := fmt.Sprintf("___HUGO_SHORTCODE_PAIRED_ALT_%d___", counter)
		placeholders[placeholder] = match
		return placeholder
	})

	// セルフクローズショートコード（{{< >}}形式）を置換
	result = hp.selfClosingShortcodeRegex.ReplaceAllStringFunc(result, func(match string) string {
		counter++
		placeholder := fmt.Sprintf("___HUGO_SHORTCODE_SELF_%d___", counter)
		placeholders[placeholder] = match
		return placeholder
	})

	// セルフクローズショートコード（{{% %}}形式）を置換
	result = hp.selfClosingShortcodeRegexAlt.ReplaceAllStringFunc(result, func(match string) string {
		counter++
		placeholder := fmt.Sprintf("___HUGO_SHORTCODE_SELF_ALT_%d___", counter)
		placeholders[placeholder] = match
		return placeholder
	})

	return result, placeholders
}

// RestoreShortcodes はプレースホルダーを元のショートコードに戻す
func (hp *HugoProcessor) RestoreShortcodes(text string, placeholders map[string]string) string {
	result := text
	for placeholder, original := range placeholders {
		result = strings.ReplaceAll(result, placeholder, original)
	}
	return result
}

// ValidateHugoMarkdown はHugoショートコードを考慮したMarkdown検証を行う
func (hp *HugoProcessor) ValidateHugoMarkdown(text string) []string {
	var issues []string

	shortcodes := hp.FindShortcodes(text)

	// ショートコードの基本的な検証
	for _, sc := range shortcodes {
		// 一般的なHugoショートコード名の検証
		if !hp.isValidShortcodeName(sc.Name) {
			issues = append(issues, "Invalid shortcode name: "+sc.Name)
		}

		// ペアードショートコードの内容検証
		if sc.Type == "paired" {
			if strings.TrimSpace(sc.Content) == "" {
				issues = append(issues, "Empty paired shortcode: "+sc.Name)
			}
		}
	}

	// ショートコードを除いた部分の基本的なMarkdown検証
	preserved, placeholders := hp.PreserveShortcodes(text)
	markdownIssues := hp.validateBasicMarkdown(preserved)
	issues = append(issues, markdownIssues...)

	// プレースホルダーを元に戻す（ログ用）
	_ = hp.RestoreShortcodes(preserved, placeholders)

	return issues
}

// isValidShortcodeName はショートコード名が有効かどうかをチェック
func (hp *HugoProcessor) isValidShortcodeName(name string) bool {
	// 基本的な命名規則をチェック
	if len(name) == 0 {
		return false
	}

	// 英数字、ハイフン、アンダースコアのみ許可
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validName.MatchString(name)
}

// validateBasicMarkdown は基本的なMarkdown構文の検証を行う
func (hp *HugoProcessor) validateBasicMarkdown(text string) []string {
	var issues []string

	lines := strings.Split(text, "\n")
	inCodeBlock := false
	codeBlockFence := ""

	for i, line := range lines {
		lineNum := i + 1

		// コードブロックの処理
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			if !inCodeBlock {
				inCodeBlock = true
				if strings.HasPrefix(line, "```") {
					codeBlockFence = "```"
				} else {
					codeBlockFence = "~~~"
				}
			} else if strings.HasPrefix(line, codeBlockFence) {
				inCodeBlock = false
				codeBlockFence = ""
			}
			continue
		}

		// コードブロック内は検証をスキップ
		if inCodeBlock {
			continue
		}

		// リンクの基本的な検証
		if strings.Contains(line, "](") {
			// Markdownリンクの基本的な形式チェック
			linkRegex := regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)`)
			matches := linkRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					linkText := match[1]
					linkURL := match[2]
					if strings.TrimSpace(linkText) == "" {
						issues = append(issues, fmt.Sprintf("Line %d: Empty link text", lineNum))
					}
					if strings.TrimSpace(linkURL) == "" {
						issues = append(issues, fmt.Sprintf("Line %d: Empty link URL", lineNum))
					}
				}
			}
		}
	}

	// 未閉じのコードブロックをチェック
	if inCodeBlock {
		issues = append(issues, "Unclosed code block")
	}

	return issues
}
