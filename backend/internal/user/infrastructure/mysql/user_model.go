// Package mysql は user モジュールのリポジトリを GORM + MySQL で実装する。
package mysql

import "time"

// UserModel が gorm.Model を埋め込まないのは、主キーが ULID で論理削除も
// 持たないため。
type UserModel struct {
	ID        string `gorm:"column:id;primaryKey;type:char(26)"`
	Name      string `gorm:"column:name;type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName は GORM 既定の複数形化（user_models）を避けて users を対象にする。
func (UserModel) TableName() string {
	return "users"
}
