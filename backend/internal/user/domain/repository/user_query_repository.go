package repository

import (
	"context"
	"errors"

	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
)

// ErrUserNotFound には、呼び出し側が errors.Is で分岐できるよう、インフラ層
// 固有の not-found エラーを変換して返す。
var ErrUserNotFound = errors.New("user not found")

type UserQueryRepository interface {
	FindByID(ctx context.Context, id model.UserID) (*model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
}
