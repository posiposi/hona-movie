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

**すべての開発コマンド(build/test/lint/migrate)はコンテナ内で実行する。ホストに Go/MySQL/Node を直接入れない。** `docker compose` で `api`(Go)/`db`(mysql:8)/`web`(Node) を定義。コマンドは `docker compose run --rm api ...` の形に統一。

- **Go は最新 stable**。`go.mod` の `go` ディレクティブと Docker イメージタグを一致させる。
- **alpine イメージは使わない**。build は `golang:<latest-stable>`、runtime は `provided.al2023` もしくは Debian系 distroless。web は `node:<lts>`。
- **Lambda 成果物**: 開発コンテナ内で `GOOS=linux GOARCH=arm64` でビルド（`provided.al2023` 想定）。
- ローカル動作確認は `docker compose up`(db/api/web) → スキーマ適用 → 検索(TMDBプロキシ)→視聴登録→比率→感想→一覧の導線を通す。

## インフラ / IaC

Terraform で全リソース管理。コスト圧縮方針が明確なので逸脱しないこと:

- **Function URL 保護**: `AWS_IAM` + CloudFront OAC（CloudFront 経由のみ許可、直叩きは 403）。OAC for Lambda の POST/PUT ボディ署名挙動は実装時に検証が必要。
- **NAT はマネージド NAT Gateway を使わない**: fck-nat(t4g.nano, ~$3/月)の NAT インスタンス。単一AZ・冗長なし。用途は TMDB への outbound のみ。
- **RDS Proxy は不要**: 単一ユーザーで同時実行ほぼ1。Lambda `reserved_concurrency` を小さく(例5)、`SetMaxOpenConns(2)`・短い `ConnMaxIdleTime` で対応。
- S3 アクセスは Gateway VPC Endpoint 経由（NAT を通さない）。シークレット(TMDB_TOKEN/DB_DSN/AUTH_TOKEN)は SSM Parameter Store SecureString。

## コーディング方針

- **コメントに頼らず、コード自体から振る舞いを読み取れるように記述する**。自明な処理を説明するだけのコメントは書かない。コメントを残すのは *why*（設計判断の理由）や TODO など、コードからは読み取れない情報に限る。

## エージェント運用ルール

- **PR作成完了後はすべてのサブエージェント・agent teamsを `TaskStop` で終了すること**。`SendMessage` での終了指示では停止しないため、必ず `TaskStop` を使用する。不要なエージェントを起動したまま残さない。

## 遵守事項

- **TMDBアトリビューション必須**: フロントのフッター等に「This product uses the TMDB API but is not endorsed or certified by TMDB.」＋TMDBロゴを掲示。TMDB は非商用・自己利用の範囲に留める。
- フロントは同一オリジン `/api` を呼ぶ（CORS不要、Cookie自動送信）。作品画像は CloudFront `/images/*` 経由で表示（TMDB CDN を直接参照しない）。
