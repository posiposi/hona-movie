---
paths:
  - "backend/**/*.go"
---

# Go Implementation Style

Google Go Style Guide (https://google.github.io/styleguide/go/) に準拠する。

## Naming

- MixedCaps/mixedCaps を使用。アンダースコアは使わない
- パッケージ名は小文字のみ、短く簡潔に。`util`/`common`/`helper` 等の汎用名は禁止
- 変数名はスコープが小さいほど短く、大きいほど説明的に
- 関数名に不要な `Get` プレフィックスを付けない
- レシーバ名は1-2文字の短い略語で統一
- インターフェースは小さく設計し、消費者側が定義する
- 定数は MixedCaps。ALL_CAPS は使わない

## Error Handling

- `error` 型は常に最後の戻り値
- エラー文字列は小文字開始、末尾句読点なし
- エラーを `_` で破棄しない
- 早期リターンで処理し `else` ブロックを避ける
- 判別には `errors.Is()` / `errors.As()` を使用（文字列パース禁止）
- ラップ形式: `fmt.Errorf("[context]: %w", err)`
- エラーをログに書くかは呼び出し側に委ねる

## Comments

- エクスポートされたトップレベル名には必ず doc コメントを付ける
- doc コメントはシンボル名で始め、完全な文で記述
- *why* を説明し、*what* の繰り返しは避ける
- unexported なコードへのコメントは必要最低限に留める

## Package Design

- 汎用パッケージ (`util`/`common`/`helper`) は作らない
- 実際に必要になるまでインターフェースを作成しない
- import 順: 標準ライブラリ → プロジェクト内 → サードパーティ
- blank import (`import _`) は main パッケージのみ

## Code Format

- `gofmt` の出力に準拠する（必須）
- **コミット前に必ず `gofmt -w` を対象ファイルに実行する**（CIのlintで検出されるため、事前に修正すること）
- 外部パッケージの構造体リテラルではフィールド名を必ず指定する
- ゼロ値フィールドは明確さを損なわない限り省略可
- 複雑な条件式はブール変数に抽出してから使用
