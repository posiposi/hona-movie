#!/bin/bash
# 統合テストは開発用スキーマとは別のスキーマに対して実行し、
# テストのクリーンアップ漏れが開発データを壊さないようにする。
#
# initdb.d 配下の .sql は環境変数を展開しないため、compose の
# MYSQL_DATABASE / MYSQL_USER と定義がずれないよう .sh で組み立てる。
set -euo pipefail

mysql --protocol=socket -uroot -p"${MYSQL_ROOT_PASSWORD}" <<-EOSQL
	CREATE DATABASE IF NOT EXISTS \`${MYSQL_DATABASE}_test\`
	    DEFAULT CHARACTER SET utf8mb4
	    DEFAULT COLLATE utf8mb4_0900_ai_ci;
	GRANT ALL PRIVILEGES ON \`${MYSQL_DATABASE}_test\`.* TO '${MYSQL_USER}'@'%';
	FLUSH PRIVILEGES;
EOSQL
