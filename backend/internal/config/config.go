// Package config は環境変数からアプリケーション設定を読み込む。
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// 単一ユーザー運用のため Lambda の同時実行はほぼ 1 であり、RDS Proxy を置かずに
// 小さなプールと短いアイドル時間で RDS の接続数を抑える。
const (
	defaultMaxOpenConns    = 2
	defaultMaxIdleConns    = 2
	defaultConnMaxLifetime = 5 * time.Minute
	defaultConnMaxIdleTime = time.Minute
)

// DB は DB 接続とコネクションプールの設定を保持する。
type DB struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// LoadDB は DB_DSN と DB_* のプール設定を環境変数から読み込む。
func LoadDB() (DB, error) {
	return loadDB("DB_DSN")
}

// LoadTestDB は統合テスト用に TEST_DB_DSN から設定を読み込む。開発用スキーマを
// 壊さないよう、テストは常にこちらの接続先を使う。
func LoadTestDB() (DB, error) {
	return loadDB("TEST_DB_DSN")
}

func loadDB(dsnKey string) (DB, error) {
	dsn := os.Getenv(dsnKey)
	if dsn == "" {
		return DB{}, fmt.Errorf("%s is not set", dsnKey)
	}

	maxOpenConns, err := intFromEnv("DB_MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return DB{}, err
	}
	maxIdleConns, err := intFromEnv("DB_MAX_IDLE_CONNS", defaultMaxIdleConns)
	if err != nil {
		return DB{}, err
	}
	connMaxLifetime, err := durationFromEnv("DB_CONN_MAX_LIFETIME", defaultConnMaxLifetime)
	if err != nil {
		return DB{}, err
	}
	connMaxIdleTime, err := durationFromEnv("DB_CONN_MAX_IDLE_TIME", defaultConnMaxIdleTime)
	if err != nil {
		return DB{}, err
	}

	return DB{
		DSN:             dsn,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
	}, nil
}

func intFromEnv(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	if value < 1 {
		return 0, fmt.Errorf("%s must be positive, got %d", key, value)
	}
	return value, nil
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	if value <= 0 {
		return 0, errors.New(key + " must be positive")
	}
	return value, nil
}
