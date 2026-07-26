package mysql

import (
	"context"
	"errors"
	"fmt"

	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
	"github.com/posiposi/hona-movie/backend/internal/user/domain/repository"

	"gorm.io/gorm"
)

// UserQueryRepository は users テーブルからの読み取りを担う。
type UserQueryRepository struct {
	db *gorm.DB
}

// NewUserQueryRepository はユーザーの読み取りリポジトリを生成する。
func NewUserQueryRepository(db *gorm.DB) repository.UserQueryRepository {
	return &UserQueryRepository{db: db}
}

func (r *UserQueryRepository) FindByID(ctx context.Context, id model.UserID) (*model.User, error) {
	var row UserModel
	if err := r.db.WithContext(ctx).Take(&row, "id = ?", id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find user %s: %w", id.Value(), repository.ErrUserNotFound)
		}
		return nil, fmt.Errorf("find user %s: %w", id.Value(), err)
	}

	user, err := toDomainUser(row)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserQueryRepository) FindAll(ctx context.Context) ([]model.User, error) {
	var rows []UserModel
	// 主キーが ULID なので id 昇順がそのまま生成順になる。
	if err := r.db.WithContext(ctx).Order("id").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	}

	users := make([]model.User, 0, len(rows))
	for _, row := range rows {
		user, err := toDomainUser(row)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func toDomainUser(row UserModel) (model.User, error) {
	id, err := model.ParseUserID(row.ID)
	if err != nil {
		return model.User{}, fmt.Errorf("convert user id %s: %w", row.ID, err)
	}
	name, err := model.NewUserName(row.Name)
	if err != nil {
		return model.User{}, fmt.Errorf("convert user name of %s: %w", row.ID, err)
	}
	return model.ReconstructUser(id, name, row.CreatedAt, row.UpdatedAt), nil
}
