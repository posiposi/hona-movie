---
paths:
  - "backend/**/*.go"
---

# Go DDD Architecture Rules

本プロジェクトのバックエンドはモジュラーモノリス + ドメイン駆動設計（DDD）+ CQRS パターンに従う。

## Layer Structure (Go)

各モジュール（境界づけられたコンテキスト）に DDD 4層を適用する。
具体的なディレクトリ構成は詳細設計で確定するが、層の責務と依存方向は以下に従う。

| Layer | 責務 |
|-------|------|
| ドメイン層 | ビジネスロジックの核。エンティティ・値オブジェクト・集約・リポジトリ抽象・ドメインサービス |
| アプリケーション層 | ユースケース（制御フロー）。DTO変換もここで行う |
| インフラ層 | 技術的実装詳細。リポジトリ実装・外部API連携・DB接続 |
| インターフェース層 | HTTP ハンドラ・ルーティング |

## File Naming Convention (Go)

| Layer | Category | Naming Pattern | Example |
|-------|----------|----------------|---------|
| domain/model | Entity | `{entity}.go` | `movie.go`, `user.go`, `watched_movie.go` |
| domain/model | Value Object | `{name}.go` | `movie_id.go`, `email.go`, `actor_ratio.go` |
| domain/model | Enum/定数 | `{name}.go` | `role.go`, `status.go` |
| domain/model | 共通基盤 | `{name}.go` | `error.go`, `id.go` |
| domain/repository | Port (Query) | `{entity}_query_port.go` | `movie_query_port.go` |
| domain/repository | Port (Command) | `{entity}_command_port.go` | `watched_movie_command_port.go` |
| infrastructure/persistence | Port Implementation | `{entity}_{query\|command}_repository.go` | `movie_query_repository.go` |
| infrastructure/persistence | テスト | `{entity}_{query\|command}_repository_test.go` | `movie_query_repository_test.go` |

### DTO変換の方針

- DTO（外部入出力の変換）はアプリケーション層（ユースケース）で行う。独自のmapperメソッドは不要
- Update操作はドメインモデル内にCommand構造体を定義して責務を寄せる

```go
type CommandUpdateUser struct {
    DisplayName string
}
```

## Query/Command Separation (CQRS)

Port（インターフェース）は原則として Query と Command に分離する。

### Query Port

- 読み取り専用の操作を定義する
- データを変更しない（副作用なし）
- `Find`, `Get`, `List` 等の動詞を使用する

```go
type MovieQueryPort interface {
    FindByID(ctx context.Context, id MovieID) (*Movie, error)
    FindByTmdbID(ctx context.Context, tmdbID TmdbMovieID) (*Movie, error)
    SearchByTitle(ctx context.Context, title string) ([]*Movie, error)
}
```

### Command Port

- 書き込み・状態変更の操作を定義する
- `Save`, `Delete`, `Update` 等の動詞を使用する

```go
type WatchedMovieCommandPort interface {
    Save(ctx context.Context, record *WatchedMovie) (*WatchedMovie, error)
    Delete(ctx context.Context, id WatchedMovieID) error
}
```

### Mixed Port (例外的)

- 集約が小さく Query/Command 分離が過剰な場合のみ許容する
- 例: Actor のように参照と upsert キャッシュが密結合な場合

```go
type ActorPort interface {
    Save(ctx context.Context, actor *Actor) (*Actor, error)
    FindByID(ctx context.Context, id ActorID) (*Actor, error)
    FindByTmdbID(ctx context.Context, tmdbID TmdbPersonID) (*Actor, error)
}
```

## Repository

- Port インターフェースはドメイン層に配置する
- Port 実装（リポジトリ）はインフラ層に配置する
- DB行→ドメインエンティティの変換はリポジトリ実装内のprivate関数で行う（独立したmapperは不要）
- 独立したmapper層は設けない（現在のサービス規模では不要）

## Dependency Rule

- Domain 層は他の層に依存しない（リポジトリはインターフェースのみ）
- Application 層は Domain 層に依存する
- Infrastructure 層は Domain 層に依存する（リポジトリインターフェースを実装）
- Infrastructure 層の詳細が Domain/Application に漏れてはならない
