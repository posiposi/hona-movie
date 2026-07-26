package migrations

import "embed"

// FS はマイグレーション SQL をバイナリに埋め込む。本番の RDS へは踏み台経由で
// バイナリ単体を持ち込んで適用するため、migrations ディレクトリの同期を不要にする。
//
//go:embed *.sql
var FS embed.FS
