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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ymotongpoo/grh"
	"gopkg.in/yaml.v3"
)

// CLIOptions はコマンドラインオプションを表す構造体
type CLIOptions struct {
	RulesYAML bool
	RulesJSON bool
	Rules     string
	Verify    bool
	Stdout    bool
	Diff      bool
	Replace   bool
	Files     []string
}

// Statistics は処理統計を表す構造体
type Statistics struct {
	FilesProcessed   int
	FilesModified    int
	TotalReplacements int
	FileStats        []FileStatistics
}

// FileStatistics はファイル毎の統計を表す構造体
type FileStatistics struct {
	FilePath     string
	Replacements int
	Modified     bool
}

func main() {
	var opts CLIOptions

	// コマンドラインフラグの定義
	flag.BoolVar(&opts.RulesYAML, "rules-yaml", false, "読み込んだルールを標準出力にYAML形式で表示する")
	flag.BoolVar(&opts.RulesJSON, "rules-json", false, "読み込んだルールを標準出力にJSON形式で表示する")
	flag.StringVar(&opts.Rules, "rules", "", "grhコマンドを実行する際のルールファイルを指定する")
	flag.BoolVar(&opts.Verify, "verify", false, "指定したファイルがMarkdownとして正しいか確認する")
	flag.BoolVar(&opts.Stdout, "stdout", false, "指定したファイルをルールファイルに基づいて置換した結果を標準出力に表示する")
	flag.BoolVar(&opts.Diff, "diff", false, "指定したファイルとそれをルールファイルに基づいて置換した結果をUnified diff形式で出力する")
	flag.BoolVar(&opts.Replace, "r", false, "指定したファイルをルールファイルに基づいて置換し上書きする")
	flag.BoolVar(&opts.Replace, "replace", false, "指定したファイルをルールファイルに基づいて置換し上書きする")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] 対象ファイル [対象ファイル...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// 残りの引数をファイルリストとして取得
	opts.Files = flag.Args()

	// ロガーの設定（--verifyオプション使用時はInfoレベル、それ以外はWarnレベル）
	logLevel := slog.LevelWarn
	if opts.Verify {
		logLevel = slog.LevelInfo
	}
	
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	if err := run(opts, logger); err != nil {
		logger.Error("Command failed", "error", err)
		os.Exit(1)
	}
}

// 統計表示用のテンプレート
const statsTemplate = `
処理結果:
  処理ファイル数: {{.FilesProcessed}}
  変更ファイル数: {{.FilesModified}}
  総置換回数: {{.TotalReplacements}}
{{if gt (len .FileStats) 0}}
ファイル別詳細:{{range .FileStats}}{{if .Modified}}
  {{.FilePath}}: {{.Replacements}}件の置換{{end}}{{end}}
{{end}}`

// printStatistics は統計情報を標準出力に表示する
func printStatistics(stats Statistics) error {
	tmpl, err := template.New("stats").Parse(statsTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	
	return tmpl.Execute(os.Stdout, stats)
}

func run(opts CLIOptions, logger *slog.Logger) error {
	// ルールファイルの読み込み
	var config *grh.Config
	var err error

	if opts.Rules != "" {
		// 指定されたルールファイルを読み込み
		config, err = grh.LoadConfigWithImports(opts.Rules)
		if err != nil {
			return fmt.Errorf("failed to load rules file %q: %w", opts.Rules, err)
		}
	} else {
		// デフォルトのルールファイルを検索
		ruleFile, err := grh.FindRuleFile("")
		if err != nil {
			return fmt.Errorf("failed to find rule file: %w", err)
		}
		config, err = grh.LoadConfigWithImports(ruleFile)
		if err != nil {
			return fmt.Errorf("failed to load rule file %q: %w", ruleFile, err)
		}
	}

	logger.Info("Loaded configuration", "rules_count", len(config.Rules), "source_paths", config.SourcePaths)

	// --rules-yaml オプションの処理
	if opts.RulesYAML {
		return outputRulesYAML(config)
	}

	// --rules-json オプションの処理
	if opts.RulesJSON {
		return outputRulesJSON(config)
	}

	// ファイルが指定されていない場合はエラー
	if len(opts.Files) == 0 {
		return fmt.Errorf("no files specified")
	}

	// Replacerを作成
	replacer := grh.NewReplacerWithLogger(config, logger)

	// 統計情報を初期化
	stats := Statistics{
		FilesProcessed: 0,
		FilesModified:  0,
		TotalReplacements: 0,
		FileStats: make([]FileStatistics, 0),
	}

	// 各ファイルを処理
	for _, filePath := range opts.Files {
		fileStat, err := processFile(filePath, opts, replacer, logger)
		if err != nil {
			return fmt.Errorf("failed to process file %q: %w", filePath, err)
		}
		
		stats.FilesProcessed++
		if fileStat.Modified {
			stats.FilesModified++
		}
		stats.TotalReplacements += fileStat.Replacements
		stats.FileStats = append(stats.FileStats, fileStat)
	}

	// 統計情報を表示（--verify, --rules-yaml, --rules-json以外の場合）
	if !opts.Verify && !opts.RulesYAML && !opts.RulesJSON {
		if err := printStatistics(stats); err != nil {
			logger.Warn("Failed to print statistics", "error", err)
		}
	}

	return nil
}

func outputRulesYAML(config *grh.Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	fmt.Print(string(data))
	return nil
}

func outputRulesJSON(config *grh.Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func processFile(filePath string, opts CLIOptions, replacer *grh.Replacer, logger *slog.Logger) (FileStatistics, error) {
	logger.Info("Processing file", "file_path", filePath)

	fileStat := FileStatistics{
		FilePath:     filePath,
		Replacements: 0,
		Modified:     false,
	}

	// --verify オプションの処理
	if opts.Verify {
		err := verifyMarkdown(filePath, replacer, logger)
		return fileStat, err
	}

	// ファイルを処理
	result, err := replacer.ReplaceFile(filePath)
	if err != nil {
		return fileStat, err
	}

	// 統計情報を更新
	fileStat.Replacements = len(result.Changes)
	fileStat.Modified = result.Changed

	// --stdout オプションの処理
	if opts.Stdout {
		fmt.Print(result.Result)
		return fileStat, nil
	}

	// --diff オプションの処理
	if opts.Diff {
		diff := replacer.GenerateDiff(result, filePath)
		if diff != "" {
			fmt.Print(diff)
		}
		return fileStat, nil
	}

	// --replace オプションの処理
	if opts.Replace {
		err := replacer.WriteResult(result, filePath)
		return fileStat, err
	}

	// デフォルト動作：変更があった場合のみ通知
	if result.Changed {
		logger.Info("File would be changed", 
			"file_path", filePath, 
			"changes_count", len(result.Changes))
		
		// 変更内容の概要を表示
		for _, change := range result.Changes {
			logger.Info("Rule would apply", 
				"rule_index", change.RuleIndex,
				"expected", change.Rule.Expected)
		}
	} else {
		logger.Info("No changes needed", "file_path", filePath)
	}

	return fileStat, nil
}

func verifyMarkdown(filePath string, replacer *grh.Replacer, logger *slog.Logger) error {
	// ファイルの拡張子をチェック
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".md" && ext != ".markdown" {
		logger.Warn("File is not a Markdown file", "file_path", filePath, "extension", ext)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	err = replacer.ValidateMarkdown(file)
	if err != nil {
		return fmt.Errorf("markdown validation failed: %w", err)
	}

	logger.Info("Markdown validation passed", "file_path", filePath)
	return nil
}
