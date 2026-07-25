---
name: implement-task
description: GitHub IssueからPR作成までの開発ワークフローを実行する。Issue番号を引数として受け取る。
allowed-tools: Read, Write, Edit, Glob, Grep, Bash, Skill, Agent, WebFetch, TaskCreate, TaskUpdate, TaskList, TaskGet
---

# 概要

引数で受け取ったGitHub Issue番号（以下 `<Issue番号>`）の仕様に基づき、以下のフェーズを順番に実行する。
各フェーズ間の情報連携はClaude CodeのTasks機能を使用する。

## Phase 1: 仕様取得

1. GitHub MCPサーバー（`mcp__github__issue_read`）を使用してIssue #`<Issue番号>` の内容を取得する
2. Issueのタイトル、本文、ラベル、コメントを収集する
3. TaskCreateで「仕様取得」タスクを作成する
   - subject: `Issue #<Issue番号> の仕様取得`
   - metadataにIssue仕様を格納する

## Phase 2: 調査・タスク分解

### Phase 2-1: 並列調査

以下のサブエージェントを**並列実行**する：

| サブエージェント | 目的 |
|---|---|
| `code-investigator` | Issue仕様に関連する既存コード・パターン・影響範囲を調査 |
| `log-investigator` | エラーログ・テスト結果・環境情報を調査（バグ修正Issue時） |

各サブエージェントには仕様取得タスクIDと自身のタスクIDを渡す。

### Phase 2-2: タスク分解

調査完了後、`task-decomposer`エージェントを起動し、1コミット粒度の実装タスクを作成する。

## Phase 3: 実装

### Phase 3-1: 実装ループ

タスク分解で作成された各実装タスクについて、以下のサイクルを実行する：

1. **テストコード作成**: `implementer`エージェントにテストコード作成を指示
2. **RED確認**: `implementer`エージェントにテスト実行（失敗確認）を指示
3. **実装**: `implementer`エージェントにコード実装を指示
4. **GREEN確認**: `implementer`エージェントにテスト実行（成功確認）を指示
5. **レビュー**: `code-reviewer`エージェントでレビュー実行
6. **修正**（指摘がある場合）: `review-fixer`エージェントで修正、再レビュー

### Phase 3-2: コミット

各タスク完了後、`/commit-commands:commit` でコミットする。

## Phase 4: PR作成

全タスク完了後：

1. `pr-creator`エージェントでPRタイトル・本文を生成する
2. `/commit-commands:commit-push-pr` でPRを作成する

## 注意事項

- 各フェーズの結果はTasks機能のmetadataで連携する
- 実装はすべてDockerコンテナ内で行う
- テスト・lint・ビルドの実行はDockerコンテナ内で行う
