package model

import (
	"strings"
	"unicode/utf8"

	"github.com/posiposi/hona-movie/backend/internal/kernel"
)

const ErrCodeInvalidUserName = "INVALID_USER_NAME"

// users.name が VARCHAR(255) であることに合わせる。MySQL は utf8mb4 でも文字数で
// 数えるため、Go 側もバイト数ではなくルーン数で数える。
const userNameMaxLength = 255

type UserName struct {
	value string
}

func NewUserName(value string) (UserName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return UserName{}, kernel.NewError(ErrCodeInvalidUserName, "user name must not be empty")
	}
	if utf8.RuneCountInString(trimmed) > userNameMaxLength {
		return UserName{}, kernel.NewError(ErrCodeInvalidUserName, "user name must be 255 characters or fewer")
	}
	return UserName{value: trimmed}, nil
}

func (n UserName) Value() string {
	return n.value
}

func (n UserName) Equals(other UserName) bool {
	return n.value == other.value
}
