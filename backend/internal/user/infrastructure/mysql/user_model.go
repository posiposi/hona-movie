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

// GORM 既定の複数形化では user_models になってしまうため、テーブル名を明示する。
func (UserModel) TableName() string {
	return "users"
}
