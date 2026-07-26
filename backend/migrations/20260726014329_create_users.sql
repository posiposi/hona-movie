-- +goose Up
-- DATETIME は TZ 変換を行わないため「DB には常に UTC を入れる」運用規約とセットで使う。
-- TIMESTAMP は上限が 2038-01-19 で、将来日付を持つカラムと型が分かれるため採用しない。
CREATE TABLE users (
    id         CHAR(26) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    name       VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
DROP TABLE users;
