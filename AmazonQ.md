# AmazonQ用実装指針

* 本プロジェクトはREADME.mdに記載されているツールの開発を行うプロジェクトです
* 本プロジェクトを行う際は日本語でやり取りしてください

## Go言語に関して

本プロジェクトはGo言語で実装します。あなたはGo言語のプロフェッショナルです。したがって次の方針で作業を進めてください。

* 指定されたライブラリ以外は極力標準パッケージを使用します
* デバッグログを含めてログを表示したい場合には `log/slog` パッケージを使用し、 [JSONL][] 形式で表示してください。
* ファイルに変更を加えるたびに `go build` を実行し、エラーが起きないことを確認してください。
* `go build` を実行してエラーが起きた場合は、エラーが起きないように修正してください。
* `go build` でエラーが出なくなったら `go test` を実行してエラーが出ないことを確認してください。
* `go test` でエラーが起きた場合、テストケース（ `*_test.go` のファイル）はいじらずに、実装だけを修正してください。
* タスクが終わるたびに各テストがREADME.mdにある仕様と一致しているか確認してください。
* もし対象のテストが仕様と一致していない場合、ユーザーにテストが仕様と一致していないことを知らせて、どのように対応するかを確認するようにしてください。ユーザーがテストの実装を修正してほしいと言った場合のみ、テストを修正してください。

[JSONL]: https://jsonlines.org/

## 作業の進め方に関して

開発の進め方は以下の手順で行います

1. まずプロジェクトをGitリポジトリとして初期化し、mainブランチを作ります
2. README.md を読み仕様を理解します
3. 実装に必要な依存している他の仕様を理解します。
4. 実装のために必要なタスクを洗い出しTODOリストの形式で列挙します。
5. TODOリストの内容をユーザーに確認し、実装を始めて良いか確認します。実装開始の許可が出るまでは絶対に実装を始めてはいけません。
6. 実装開始をする前にGitリポジトリに作業前のコミットをし、作業用ブランチを切ります。
7. 最初のタスクを開始します。ファイルを1つ追加する事に作業用ブランチにコミットをします。
8. 1つのタスクが終わったらTODOリストでチェックをし、ユーザーに作業用ブランチをmainブランチにマージをする確認を取ります。
9. ユーザーがマージをして良いと言ったらmainブランチにマージをします。
10. 次のタスク用の作業用ブランチを切ります
11. 以下8-10を繰り返します

なお、Gitコミットのコミットログは次のような決まりに基づいて日本語で書きます。

```shell-session
$ git commit -m "{タイプ}: {作業内容の1行要約} -m "
{作業内容の詳細}
"
```

`{タイプ}` に関しては作業内容に応じて次のラベルを使い分けてください

* add: 新規機能の追加の場合
* change: 既存機能の修正の場合
* fix: 既存機能にあったバグを修正する場合
* doc: ドキュメントの修正だけの場合

複数のタイプが混在しそうな変更を行った場合にはコミットを分けてください。