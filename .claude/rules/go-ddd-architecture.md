---
paths:
  - "backend/**/*.go"
---

# Go DDD Architecture Rules

本プロジェクトのバックエンドはモジュラーモノリス + ドメイン駆動設計（DDD）+ CQRS パターンに従う。

## Layer Structure (Go)

各モジュール（境界づけられたコンテキスト）に DDD 4層を適用する。

| Layer | 責務 |
|-------|------|
| ドメイン層 | ビジネスロジックの核。エンティティ・値オブジェクト・集約・リポジトリ抽象・ドメインサービス |
| アプリケーション層 | ユースケース（制御フロー）。DTO変換もここで行う |
| インフラ層 | 技術的実装詳細。リポジトリ実装・外部API連携・DB接続 |
| インターフェース層 | HTTP ハンドラ・ルーティング |

## Directory Structure (Go)

モジュール（境界づけられたコンテキスト）を `internal/{module}/` 単位で切り、その配下に DDD 4層を置く。

```
backend/
├── cmd/                                # エントリポイント（全モジュールの配線点）
├── migrations/                         # 物理DBが単一のためモジュールに含めず一元管理
└── internal/
    ├── kernel/                         # 共有カーネル。ID基底型・ドメインエラー
    ├── config/                         # 技術基盤。環境変数 → 設定
    ├── db/                             # 技術基盤。DB接続
    └── {module}/                       # 例: user, movie, watched_movie, review, auth, admin
        ├── domain/
        │   ├── model/
        │   └── repository/             # リポジトリインターフェース（抽象）
        ├── application/
        ├── infrastructure/
        │   └── mysql/                  # リポジトリ実装（実体）
        └── interfaces/
```

- モジュール間はドメイン層のインターフェース経由で参照し、他モジュールのインフラ層には依存しない
- `kernel` は全モジュールが依存してよい唯一のドメイン共有物。ここに業務ロジックは置かない
- 汎用パッケージ（`util` / `common` / `helper`）は作らない

## File Naming Convention (Go)

| Layer | Category | Naming Pattern | Example |
|-------|----------|----------------|---------|
| domain/model | Entity | `{entity}.go` | `movie.go`, `user.go`, `watched_movie.go` |
| domain/model | Value Object | `{name}.go` | `movie_id.go`, `email.go`, `actor_ratio.go` |
| domain/model | Enum/定数 | `{name}.go` | `role.go`, `status.go` |
| domain/model | 共通基盤 | `{name}.go` | `error.go`, `id.go` |
| domain/repository | Repository Interface (Query) | `{entity}_query_repository.go` | `movie_query_repository.go` |
| domain/repository | Repository Interface (Command) | `{entity}_command_repository.go` | `watched_movie_command_repository.go` |
| infrastructure/mysql | Repository Implementation | `{entity}_{query\|command}_repository.go` | `movie_query_repository.go` |
| infrastructure/mysql | GORM モデル | `{entity}_model.go` | `movie_model.go` |
| infrastructure/mysql | テスト | `{entity}_{query\|command}_repository_test.go` | `movie_query_repository_test.go` |

インターフェースと実装はファイル名・型名とも同一にし、パッケージ名（`repository` / `mysql`）で区別する。

```go
// internal/user/domain/repository/user_query_repository.go
package repository
type UserQueryRepository interface { ... }

// internal/user/infrastructure/mysql/user_query_repository.go
package mysql
type UserQueryRepository struct { db *gorm.DB }

func NewUserQueryRepository(db *gorm.DB) repository.UserQueryRepository { ... }
```

### DTO変換の方針

- DTO（外部入出力の変換）はアプリケーション層（ユースケース）で行う。独自のmapperメソッドは不要
- Update操作はドメインモデル内にCommand構造体を定義して責務を寄せる

```go
type CommandUpdateUser struct {
    DisplayName string
}
```

## Query/Command Separation (CQRS)

リポジトリインターフェースは原則として Query と Command に分離する。

### Query Repository

- 読み取り専用の操作を定義する
- データを変更しない（副作用なし）
- `Find`, `Get`, `List` 等の動詞を使用する

```go
type MovieQueryRepository interface {
    FindByID(ctx context.Context, id MovieID) (*Movie, error)
    FindByTmdbID(ctx context.Context, tmdbID TmdbMovieID) (*Movie, error)
    SearchByTitle(ctx context.Context, title string) ([]*Movie, error)
}
```

### Command Repository

- 書き込み・状態変更の操作を定義する
- `Save`, `Delete`, `Update` 等の動詞を使用する

```go
type WatchedMovieCommandRepository interface {
    Save(ctx context.Context, record *WatchedMovie) (*WatchedMovie, error)
    Delete(ctx context.Context, id WatchedMovieID) error
}
```

### Mixed Repository (例外的)

- 集約が小さく Query/Command 分離が過剰な場合のみ許容する
- 例: Actor のように参照と upsert キャッシュが密結合な場合

```go
type ActorRepository interface {
    Save(ctx context.Context, actor *Actor) (*Actor, error)
    FindByID(ctx context.Context, id ActorID) (*Actor, error)
    FindByTmdbID(ctx context.Context, tmdbID TmdbPersonID) (*Actor, error)
}
```

## Repository

- リポジトリインターフェースはドメイン層（`domain/repository`）に配置する
- 実装はインフラ層（`infrastructure/mysql`）に配置する
- DB行→ドメインエンティティの変換はリポジトリ実装内のprivate関数で行う（独立したmapperは不要）
- 独立したmapper層は設けない（現在のサービス規模では不要）
- GORM のタグ付き構造体はインフラ層の関心事であり、ドメインエンティティとは別型にする

## Dependency Rule

- Domain 層は他の層に依存しない（リポジトリはインターフェースのみ）
- Application 層は Domain 層に依存する
- Infrastructure 層は Domain 層に依存する（リポジトリインターフェースを実装）
- Infrastructure 層の詳細が Domain/Application に漏れてはならない
