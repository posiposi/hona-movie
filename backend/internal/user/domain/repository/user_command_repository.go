package repository

import (
	"context"

	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
)

type UserCommandRepository interface {
	Save(ctx context.Context, user model.User) error
	Delete(ctx context.Context, id model.UserID) error
}
