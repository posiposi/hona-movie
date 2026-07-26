//go:build integration

package mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/posiposi/hona-movie/backend/internal/config"
	"github.com/posiposi/hona-movie/backend/internal/db"
	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
	"github.com/posiposi/hona-movie/backend/migrations"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

// テストデータはこのプレフィックスで識別し、開発用データと混ざらないようにする。
const testDataPrefix = "user-repo-test-"

var testDB *gorm.DB

func TestMain(m *testing.M) {
	code, err := runTests(m)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

func runTests(m *testing.M) (int, error) {
	cfg, err := config.LoadTestDB()
	if err != nil {
		return 0, err
	}
	if err := migrateTestSchema(cfg); err != nil {
		return 0, err
	}

	ctx := context.Background()
	gormDB, err := db.Open(ctx, cfg)
	if err != nil {
		return 0, err
	}
	defer db.Close(gormDB)

	testDB = gormDB
	return m.Run(), nil
}

// migrateTestSchema が接続先の `_test` サフィックスを確認してから適用するのは、
// 誤って開発用スキーマを対象にした場合にテストが破壊してしまうため。
func migrateTestSchema(cfg config.DB) error {
	sqlDB, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return fmt.Errorf("open test database: %w", err)
	}
	defer sqlDB.Close()

	var schema string
	if err := sqlDB.QueryRow("SELECT DATABASE()").Scan(&schema); err != nil {
		return fmt.Errorf("resolve current schema: %w", err)
	}
	if !strings.HasSuffix(schema, "_test") {
		return fmt.Errorf("TEST_DB_DSN points to %q, want a schema with the _test suffix", schema)
	}

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.UpContext(context.Background(), sqlDB, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func newTestUserName(t *testing.T, suffix string) model.UserName {
	t.Helper()
	name, err := model.NewUserName(testDataPrefix + suffix)
	if err != nil {
		t.Fatalf("NewUserName(%v) returned error: %v", testDataPrefix+suffix, err)
	}
	return name
}

// リポジトリ実装を介さず直接 INSERT するのは、Query 側のテストを Command 側の
// 実装から独立させるため。
func insertTestUser(t *testing.T, id model.UserID, name model.UserName, at time.Time) {
	t.Helper()
	err := testDB.Exec(
		"INSERT INTO users (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)",
		id.Value(), name.Value(), at, at,
	).Error
	if err != nil {
		t.Fatalf("insert test user(%v) returned error: %v", id.Value(), err)
	}
	cleanupUsers(t, id)
}

func cleanupUsers(t *testing.T, ids ...model.UserID) {
	t.Helper()
	t.Cleanup(func() {
		values := make([]string, 0, len(ids))
		for _, id := range ids {
			values = append(values, id.Value())
		}
		if err := testDB.Exec("DELETE FROM users WHERE id IN ?", values).Error; err != nil {
			t.Errorf("cleanup users(%v) returned error: %v", values, err)
		}
	})
}
