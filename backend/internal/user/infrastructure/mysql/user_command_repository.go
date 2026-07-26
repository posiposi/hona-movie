package mysql

import (
	"context"
	"fmt"

	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
	"github.com/posiposi/hona-movie/backend/internal/user/domain/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserCommandRepository は users テーブルへの書き込みを担う。
type UserCommandRepository struct {
	db *gorm.DB
}

// NewUserCommandRepository はユーザーの書き込みリポジトリを生成する。
func NewUserCommandRepository(db *gorm.DB) repository.UserCommandRepository {
	return &UserCommandRepository{db: db}
}

// Save は新規作成と更新の双方を UPSERT で扱う。created_at / updated_at は
// GORM に採番させるため、ドメインの値は永続化に持ち込まない。
func (r *UserCommandRepository) Save(ctx context.Context, user model.User) error {
	row := UserModel{
		ID:   user.ID().Value(),
		Name: user.Name().Value(),
	}
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&row).Error
	if err != nil {
		return fmt.Errorf("save user %s: %w", row.ID, err)
	}
	return nil
}

func (r *UserCommandRepository) Delete(ctx context.Context, id model.UserID) error {
	result := r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id.Value())
	if result.Error != nil {
		return fmt.Errorf("delete user %s: %w", id.Value(), result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete user %s: %w", id.Value(), repository.ErrUserNotFound)
	}
	return nil
}
