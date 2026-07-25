---
name: linter-execute
description: テスト通過後のlint確認・修正手順を定義するスキル。Dockerコンテナ内でlintコマンドを実行する。TDD実装フロー内でテストPASS後に使用する。
user-invocable: false
allowed-tools: Bash
---

# Linter実行スキル

## 実行タイミング

テストがPASSした後、コミット前にlint確認を行う。

## バックエンドのlint実行

apiコンテナ内で下記コマンドを順に実行する。

```bash
# 1. フォーマット確認・修正
docker compose run --rm api gofmt -w .

# 2. vet（静的解析）
docker compose run --rm api go vet ./...
```

## フロントエンドのlint実行

webコンテナ内で下記コマンドを順に実行する（フロントエンド環境構築後）。

```bash
# 1. lint確認
docker compose run --rm web npm run lint

# 2. 型チェック
docker compose run --rm web npm run typecheck

# 3. フォーマット確認
docker compose run --rm web npm run format:check
```

lintエラーがある場合はwebコンテナ内で自動修正を実行する。

```bash
docker compose run --rm web npm run format
```

## 制約事項

- lint/formatコマンドは**必ず**Dockerコンテナ内で実行する
- バックエンドは`gofmt -w` → `go vet`の順で実行する
- lintエラーの場合はlintを無効化するアノテーション等を使用せず、根本的な問題解決を行う
