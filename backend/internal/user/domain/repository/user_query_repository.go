package repository

import (
	"context"
	"errors"

	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
)

// ErrUserNotFound は該当するユーザーが存在しないことを表す。呼び出し側が
// errors.Is で分岐できるよう、インフラ層固有のエラーはこれに変換する。
var ErrUserNotFound = errors.New("user not found")

// UserQueryRepository はユーザーの読み取り操作を表すリポジトリインターフェース。
type UserQueryRepository interface {
	FindByID(ctx context.Context, id model.UserID) (*model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
}
