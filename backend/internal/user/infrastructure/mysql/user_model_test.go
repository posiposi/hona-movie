//go:build integration

package mysql_test

import (
	"testing"

	usermysql "github.com/posiposi/hona-movie/backend/internal/user/infrastructure/mysql"

	"gorm.io/gorm"
)

// TestUserModelMatchesSchema は GORM モデルと実スキーマのカラム集合を双方向に突き合わせる。
// マイグレーションでカラムを増減させたときにモデルの更新漏れを検出するのが狙い。
func TestUserModelMatchesSchema(t *testing.T) {
	statement := &gorm.Statement{DB: testDB}
	if err := statement.Parse(&usermysql.UserModel{}); err != nil {
		t.Fatalf("Parse(UserModel) returned error: %v", err)
	}

	modelColumns := make(map[string]bool, len(statement.Schema.DBNames))
	for _, dbName := range statement.Schema.DBNames {
		modelColumns[dbName] = true
	}

	columnTypes, err := testDB.Migrator().ColumnTypes(&usermysql.UserModel{})
	if err != nil {
		t.Fatalf("ColumnTypes(UserModel) returned error: %v", err)
	}
	schemaColumns := make(map[string]bool, len(columnTypes))
	for _, columnType := range columnTypes {
		schemaColumns[columnType.Name()] = true
	}

	for column := range modelColumns {
		if !schemaColumns[column] {
			t.Errorf("UserModel declares column %v, want it to exist in table %v", column, statement.Schema.Table)
		}
	}
	for column := range schemaColumns {
		if !modelColumns[column] {
			t.Errorf("table %v has column %v, want it to be declared in UserModel", statement.Schema.Table, column)
		}
	}
}
