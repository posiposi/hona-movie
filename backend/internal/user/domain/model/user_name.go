package model

import (
	"strings"
	"unicode/utf8"

	"github.com/posiposi/hona-movie/backend/internal/kernel"
)

const ErrCodeInvalidUserName = "INVALID_USER_NAME"

// userNameMaxLength は users.name が VARCHAR(255) であることに合わせた上限。
// バイト数ではなく文字数で数えるため、マルチバイト文字も 255 文字まで許容する。
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
