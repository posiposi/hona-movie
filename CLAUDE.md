# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

**honamovie** — 自分が視聴した映画を記録する個人用(単一ユーザー)Webアプリ。グリーンフィールド新規構築。映画検索・基本情報は TMDB API を使い、視聴記録・感想・「お気に入り俳優の出演比率」は自前DBに保存する。詳細な設計判断は `project_plan.md` に集約されている（実装前に必ず参照すること）。

## リポジトリ構成

トップレベルでフロント/バック/インフラを分離した3ディレクトリ構成。各ディレクトリ内部の詳細構成は「Issue 起票時の詳細設計で確定」する方針のため、現段階では空（`.gitkeep` のみ）。

```
frontend/   # React SPA(Vite)。エンドユーザー向け(CloudFront配信)と Admin管理画面(非公開)をビルド/成果物分離
backend/    # Go Lambdalith。モジュラーモノリス + DDD + CQRS
infra/      # Terraform
```

`backend/` の内部構成はモジュール（境界づけられたコンテキスト）単位で切る。詳細は `.claude/rules/go-ddd-architecture.md` を参照。

```
backend/
├── cmd/                  # エントリポイント（全モジュールの配線点）
├── migrations/           # 物理DBが単一のためモジュールに含めず一元管理
└── internal/
    ├── kernel/           # 共有カーネル。ID基底型・ドメインエラー。業務ロジックは置かない
    ├── config/           # 技術基盤。環境変数 → 設定
    ├── db/               # 技術基盤。GORM 接続
    └── {module}/         # user, movie, watched_movie, review, actor_ratio, auth, admin
        ├── domain/{model,repository}
        ├── application/
        ├── infrastructure/mysql/
        └── interfaces/
```

## アーキテクチャ（確定済みの重要方針）

- **単一ドメイン配信**: CloudFront 単一ドメインで SPA・画像・API を配信し CORS を回避。同一オリジンなので HttpOnly Cookie がそのまま届く。`default→S3(SPA)` / `/images/*→S3(画像)` / `/api/*→Lambda Function URL`。
- **バックエンドは Lambdalith**: 単一 Go バイナリを Lambda + Function URL(API Gateway/ALB なし)で公開。ローカルでは同一コードを通常の HTTP サーバとして起動し、**ローカルと Lambda を同一実装で**動かす（実行環境で分岐）。
- **モジュラーモノリス + DDD + CQRS**: 単一デプロイを維持しつつ、境界づけられたコンテキスト(映画/視聴記録/感想/出演比率/認証、および別ドメインの Admin)ごとにモジュール分割。各モジュールに DDD 4層(ドメイン/アプリケーション/インフラ/インターフェース)を適用しドメイン層に業務ロジックを集約。Port（リポジトリインターフェース）は原則 Query と Command に分離する（CQRS）。
- **Admin ドメインは非公開**: Admin の HTTP 公開面は公開 CloudFront/Function URL の `/api/*` には出さない。Admin 画面も CloudFront/公開S3 では配信せずローカル/内部のみ。エンドユーザー向けとビルド/成果物を分離する。
- **TMDB は Lambda 経由の BFF プロキシ**: API トークンをフロントに露出させない。タイムアウト・リトライ・レート制限(429, ~40req/s)対応。作品詳細取得時はキャストも併せて取得。Lambda は VPC 内なので TMDB へは NAT インスタンス(fck-nat)経由で出る。
- **画像は S3 キャッシュ**: 視聴登録時に TMDB CDN からポスターを1回だけ取得→S3 アップロード→**S3キーを RDS に保存**（既存キーがあればスキップ＝冪等）。以降の表示は CloudFront `/images/*` 経由。S3 アクセスは Gateway VPC Endpoint 経由（NAT を通さない＝無料）。TMDB CDN URL の直接参照はしない。
- **認証**: 共有シークレット方式で検証し `HttpOnly; Secure; SameSite` Cookie を発行。将来のマルチユーザー化(Cognito 等)に差し替え可能な形で認証モジュールに隠蔽する。

## データモデル方針

- `users` テーブルを最初から作成（将来のマルチユーザー化に備える）。当面は自分1行運用。**ユーザー固有データ(`watched_movies`/`reviews`/`movie_actor_ratios`)には最初から `user_id` を持たせる**。
- `movies`/`actors` は**ユーザー横断の共有マスタ兼TMDBキャッシュ**なので `user_id` を持たない。TMDB の movie_id/person_id を自然キーとし、基本情報はスナップショットとしてキャッシュ・`last_synced_at` で再取得管理。
- RDS(InnoDB) なので FK 制約はそのまま利用。出演比率の合計100%制約は付けない（手動入力・主観のため緩く扱う）。
- 具体的な DDL・型・制約・インデックスは詳細設計で確定する。

## 開発環境・コマンド

**すべての開発コマンド(build/test/lint/migrate)はコンテナ内で実行する。ホストに Go/MySQL/Node を直接入れない。** `docker compose` で `api`(Go)/`db`(mysql:8.4.x LTS) を定義（フロントエンドの `web`(Node) は別 Issue で追加予定）。コマンドは `docker compose run --rm api ...` の形に統一。

- **Go は最新 stable**。`go.mod` の `go` ディレクティブと Docker イメージタグを一致させる。
- **MySQL は LTS 系（8.4.x）を使う**。9.x は Innovation トラック（サポート期間が短い）なので採用しない。イメージタグはパッチまで固定する（浮動タグ `mysql:8.4` は最新パッチに追随しないため）。
- **alpine イメージは使わない**。build は `golang:<latest-stable>`、runtime は `provided.al2023` もしくは Debian系 distroless。web は `node:<lts>`。
- **Lambda 成果物**: 開発コンテナ内で `GOOS=linux GOARCH=arm64` でビルド（`provided.al2023` 想定）。
- ローカル動作確認は `docker compose up`(db/api) → スキーマ適用 → 検索(TMDBプロキシ)→視聴登録→比率→感想→一覧の導線を通す。
- **コンテナは常時起動させておく。作業後に `docker compose down` / `stop` で停止しないこと。** 起動済みのコンテナに対して `docker compose run --rm api ...` / `docker compose exec api ...` でコマンドを実行する。

## DB マイグレーション

**pressly/goose** を versioned migration で使う。goose CLI は導入せず、`backend/cmd/migrate` が `embed.FS` 経由で SQL を読む。ローカルと本番で同一実装を通し、本番へはバイナリ単体を持ち込めば足りるようにするため。

```bash
docker compose run --rm api go run ./cmd/migrate status   # 適用状況
docker compose run --rm api go run ./cmd/migrate up       # 未適用をすべて適用
docker compose run --rm api go run ./cmd/migrate down     # 1バージョンだけロールバック
```

- DSN は `docker-compose.yml` が `api` に渡す `DB_DSN` から読む。引数での指定はしない。
- **新規マイグレーションは手で作成する**（goose CLI を入れていないため）。`backend/migrations/<UTCタイムスタンプ>_<説明>.sql` に `-- +goose Up` / `-- +goose Down` の両方を書く。タイムスタンプは `date -u +%Y%m%d%H%M%S` で採る。
- **`//go:embed *.sql` は対象ファイルが0件だとビルドエラーになる。** `backend/migrations/` を空にしないこと。
- 物理DBは単一なので、モジュラーモノリスの境界に関わらず `backend/migrations/` に一元管理する。
- **主キーは ULID(`CHAR(26)`)** とし、ID カラムのみ `CHARACTER SET ascii COLLATE ascii_bin` を指定する。テーブル既定の `utf8mb4` では `CHAR(26)` が 104 bytes を占め、クラスタリングインデックスと全 FK が肥大化するため。
- **日時カラムは `DATETIME`**（`TIMESTAMP` は上限が 2038-01-19）。TZ 変換されないため、**DB には常に UTC を入れる**運用規約とセットで守る。
- 本番(VPC内RDS)への適用経路は別 Issue で設計する。Lambda ハンドラ内でのマイグレーション実行はしない（コールドスタート増加・同時実行時の競合・実行時間のリスク）。

## ORM / データアクセス

**GORM**(`gorm.io/gorm` + `gorm.io/driver/mysql`) を読み書きのみに使う。**`AutoMigrate` は使わない**（未使用カラムを削除せず、差分がレビューできず履歴も残らないため）。スキーマの正は `backend/migrations/` の SQL であり、GORM モデルはそれに追従する。

- GORM のタグ付き構造体はインフラ層の関心事なので `infrastructure/mysql/{entity}_model.go` に置き、**ドメインエンティティとは別型**にする。DB行→ドメインエンティティの変換はリポジトリ実装内の private 関数で行う（独立した mapper 層は設けない）。
- **`gorm.Model` は使わない**。主キーが ULID(`CHAR(26)`) で論理削除も持たないため、必要なカラムだけを手で宣言する。
- GORM 既定の複数形化は当てにせず、モデルに `TableName()` を実装する（`UserModel` の既定は `user_models` になる）。
- **DB 接続は `internal/db` の `Open` に集約する**。`SetMaxOpenConns(2)` 等のプール設定と `NowFunc` の UTC 固定をここで行うため、`gorm.Open` を各所で直接呼ばない。設定値は `internal/config` が環境変数から読む。
- 見つからない場合は `gorm.ErrRecordNotFound` をそのまま返さず、`domain/repository` の sentinel エラー（`ErrUserNotFound` 等）にラップして返す。呼び出し側が `errors.Is` で判別できるようにするため。
- ULID の生成は `github.com/oklog/ulid/v2`。`internal/kernel` の ID 値オブジェクトに隠蔽してある。**`ulid.Parse` ではなく `ulid.ParseStrict` を使う**（`ulid.Parse` は速度優先で文字検証を省き、不正な文字列を黙って別の値にデコードする）。

## テスト

ユニットテストと、実DBに対する統合テストの2種類。詳細な規約は `.claude/rules/go-testing.md` を参照。

```bash
docker compose run --rm api go test ./...                    # ユニットテスト
docker compose run --rm api go test -tags=integration ./...   # 統合テスト（実DB）
```

- 統合テストは `//go:build integration` タグで区別する。ファイル名に `integration` は含めない。
- **統合テストは専用スキーマ `honamovie_test` に対して実行する**。接続先は `docker-compose.yml` が `api` に渡す `TEST_DB_DSN`。開発用スキーマ(`honamovie`)を絶対に使わない（クリーンアップ漏れで開発データを壊さないため）。`TestMain` で接続先が `_test` サフィックスを持つことを検証してから走らせる。
- テスト用スキーマは `docker/mysql/initdb.d/` の init スクリプトで作成する。**`db-data` volume が既に初期化済みの場合 init スクリプトは実行されない**ので、既存環境では手動で `CREATE DATABASE` と `GRANT` が必要。
- マイグレーションの適用は `TestMain` で goose を呼んで自動化する。手動適用を前提にしない。
- テストデータは固定プレフィックス付きで作り、`t.Cleanup` で必ず削除する。
- **GORM モデルと DDL の乖離はスキーマ突合テストで検出する**（`Migrator().ColumnTypes()` とモデルのフィールド集合を**双方向**に突き合わせる）。片方向だけだと「DDLにあるがモデルにないカラム」を見逃す。CRUD の統合テストと併用する。

## インフラ / IaC

Terraform で全リソース管理。コスト圧縮方針が明確なので逸脱しないこと:

- **Function URL 保護**: `AWS_IAM` + CloudFront OAC（CloudFront 経由のみ許可、直叩きは 403）。OAC for Lambda の POST/PUT ボディ署名挙動は実装時に検証が必要。
- **NAT はマネージド NAT Gateway を使わない**: fck-nat(t4g.nano, ~$3/月)の NAT インスタンス。単一AZ・冗長なし。用途は TMDB への outbound のみ。
- **RDS Proxy は不要**: 単一ユーザーで同時実行ほぼ1。Lambda `reserved_concurrency` を小さく(例5)、`SetMaxOpenConns(2)`・短い `ConnMaxIdleTime` で対応。
- S3 アクセスは Gateway VPC Endpoint 経由（NAT を通さない）。シークレット(TMDB_TOKEN/DB_DSN/AUTH_TOKEN)は SSM Parameter Store SecureString。

## コーディング方針

- **コメントに頼らず、コード自体から振る舞いを読み取れるように記述する**。自明な処理を説明するだけのコメントは書かない。コメントを残すのは *why*（設計判断の理由）や TODO など、コードからは読み取れない情報に限る。

## エージェント運用ルール

- **全チームの作業が完了した時点で、すべてのサブエージェント・agent teams を `TaskStop` で終了すること**。PR作成まで待たない。`SendMessage` での終了指示では停止しないため、必ず `TaskStop` を使用する。不要なエージェントを起動したまま残さない。
- 並行実行しているチームがある場合は、**全チームの完了を確認してから**まとめて終了する。先に終わったチームだけを個別に止めない（他チームからの問い合わせに応答できなくなるため）。

## 遵守事項

- **TMDBアトリビューション必須**: フロントのフッター等に「This product uses the TMDB API but is not endorsed or certified by TMDB.」＋TMDBロゴを掲示。TMDB は非商用・自己利用の範囲に留める。
- フロントは同一オリジン `/api` を呼ぶ（CORS不要、Cookie自動送信）。作品画像は CloudFront `/images/*` 経由で表示（TMDB CDN を直接参照しない）。
