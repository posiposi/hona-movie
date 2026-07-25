---
name: tdd-workflow
description: テスト駆動開発（TDD）のワークフロー定義。テスト実行はDockerコンテナ内で行う。Red→Green→Refactorのサイクル手順、テストの書き方、DDD規約との統合方法を定義する。テスト実装・実行時に使用する。
user-invocable: false
allowed-tools: Read, Write, Edit, Glob, Grep, Bash, TaskUpdate, TaskGet, TaskList
---

# TDDワークフロースキル

## TDDサイクル

### 1. Red フェーズ（テストを先に書く）

#### 手順

1. テストファイルを作成する（`*_test.go`）
2. テスト対象の期待する振る舞いをテストケースとして記述する
3. テストを実行し、**失敗すること**を確認する

#### テストの書き方

```go
func TestMovieID_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   int64
        wantErr bool
    }{
        {"正の値で生成できる", 12345, false},
        {"ゼロは不正", 0, true},
        {"負の値は不正", -1, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewMovieID(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewMovieID(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
                return
            }
            if !tt.wantErr && got.Value() != tt.input {
                t.Errorf("NewMovieID(%d).Value() = %v, want %v", tt.input, got.Value(), tt.input)
            }
        })
    }
}
```

#### テスト名

- `t.Run` のテスト名は**日本語**で振る舞いを簡潔に表現する
- Table-driven tests を活用しテストケースの追加を容易にする

#### DDD層別のテスト方針

| 層 | テスト対象 | モック対象 |
|---|---|---|
| ドメイン層 | 値オブジェクト、エンティティ、集約 | なし（純粋なロジック） |
| アプリケーション層 | ユースケース | Port（Command/Query） |
| インターフェース層 | ハンドラ | ユースケース |
| インフラ層 | リポジトリ実装 | なし（テストDBに実接続） |

#### インフラ層のテスト方針

- インフラ層のテストではモックを使用せず、**テストDB**に実際に接続してアサーションを行う
- テストDBへの接続はDockerコンテナ内のテスト環境で自動的に行われる
- 各テストケースはデータの独立性を保つため、テストデータの投入・クリーンアップを各テストコードで必ず行う
- integrationテストは `//go:build integration` タグで区別する

### 2. Green フェーズ（最小限の実装）

#### 手順

1. テストが通る**最小限**のコードを実装する
2. 「動くコード」を最優先する
3. テストを実行し、**すべて通ること**を確認する

#### 原則

- 完璧なコードを書こうとしない
- テストに書かれていない機能は実装しない
- 既存のテストが壊れていないことも確認する

### 3. Refactor フェーズ（品質改善）

#### 手順

1. コードの重複を排除する
2. 命名を改善する（DDD規約に準拠）
3. 不変性、ファクトリ関数等のDDD規約への適合を確認する
4. テストを再実行し、**すべて通ること**を確認する

#### リファクタリング観点

- 単一責任の原則に違反していないか
- DDD層の責務が正しく分離されているか
- 命名がドメインの言葉を使っているか
- 不要なコメントや死んだコードがないか

## テスト実行コマンド

**重要**: テストは必ずDockerコンテナ内で実行する。

```bash
# 特定パッケージのテスト実行
docker compose run --rm api go test ./internal/domain/model/...

# 全テスト実行
docker compose run --rm api go test ./...

# integrationテスト実行
docker compose run --rm api go test -tags=integration ./...
```

## テスト失敗時の対応

1. エラーメッセージを正確に読む
2. 期待値と実際の値の差分を確認する
3. Greenフェーズの実装を修正する（テストは変更しない）
4. テストを再実行する
5. 3回修正しても通らない場合は、テストの前提条件を見直す

## テストケースの原則

- テストケースは**そのコード固有の振る舞い（ドメインロジック・バリデーション・状態遷移等）のみ**を検証する
- 言語機能そのもの（基本的な演算、型変換等）はテストしない
- コード固有のロジックが絡むケースのみテストする
  - 例: 正規化（トリム、大文字小文字統一等）が等価性比較に影響するケース
  - 例: バリデーションルール（文字数制限、フォーマット検証等）
  - 例: ファクトリ関数での入力変換ロジック

## テストのアンチパターン

- テストが実装の詳細に依存している（unexportedな関数のテスト等）
- テスト間に順序依存がある
- テストデータが共有されている（各テストは独立すべき）
- モックが多すぎる（設計の問題のサイン）
- テスト名が意味のない名前になっている
