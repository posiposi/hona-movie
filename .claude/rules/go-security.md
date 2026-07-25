---
paths:
  - "backend/**/*.go"
---

# Go Security Rules

## SQL Injection Prevention

- SQL クエリには必ずパラメータ化クエリ（プレースホルダ）を使用する
- 文字列結合や `fmt.Sprintf` で SQL を組み立てない
- sqlc 等のコード生成ツールを活用し、手書き SQL を最小限にする
- ユーザー入力をそのまま ORDER BY や LIMIT に渡さない

## Input Validation

- 外部から受け取るすべての入力（HTTP リクエスト、環境変数、ファイル）はバリデーションする
- ドメイン層の値オブジェクトで不変条件を強制し、不正な値の生成を防ぐ
- 整数オーバーフローに注意し、境界値を検証する
- パスの操作には `filepath.Clean` を使用し、パストラバーサルを防ぐ

## Authentication & Authorization

- 認証トークンやパスワードをログに出力しない
- 秘密情報（API キー、DB パスワード等）はハードコードせず環境変数で管理する
- パスワードの比較には `subtle.ConstantTimeCompare` を使用し、タイミング攻撃を防ぐ

## Error Handling for Security

- 内部エラーの詳細をクライアントに漏洩しない（スタックトレース、DB エラー等）
- エラーレスポンスは汎用メッセージを返し、詳細はサーバーログに記録する
- パニックリカバリーのミドルウェアを設置し、未処理パニックでプロセスが落ちることを防ぐ

## Dependency Management

- 依存パッケージは定期的に脆弱性をチェックする（`govulncheck`）
- 不要な依存は削除し、攻撃対象面を最小化する
- `go.sum` をバージョン管理に含め、サプライチェーン攻撃を防ぐ

## Concurrency Safety

- 共有状態へのアクセスには `sync.Mutex` または channel を使用する
- `context.Context` を適切に伝播し、タイムアウト・キャンセルを実装する
- goroutine リークを防ぐため、起動した goroutine の終了を保証する
