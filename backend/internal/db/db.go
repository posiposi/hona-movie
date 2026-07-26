// Package db は GORM による MySQL 接続を生成する。
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/posiposi/hona-movie/backend/internal/config"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const pingTimeout = 5 * time.Second

// Open は設定から GORM 接続を生成し、疎通確認を行う。
func Open(ctx context.Context, cfg config.DB) (*gorm.DB, error) {
	gormDB, err := gorm.Open(gormmysql.Open(cfg.DSN), &gorm.Config{
		// DATETIME は TZ 変換されないため、GORM が付与する created_at/updated_at も
		// UTC で揃える（「DB には常に UTC を入れる」運用規約）。
		NowFunc: func() time.Time { return time.Now().UTC() },
		// スキーマの正は backend/migrations の SQL であり、GORM には
		// 外部キー制約を作らせない。
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("resolve sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return gormDB, nil
}

// Close は GORM 接続が保持するコネクションプールを閉じる。
func Close(gormDB *gorm.DB) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("resolve sql.DB: %w", err)
	}
	return sqlDB.Close()
}
