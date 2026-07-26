package model

import "github.com/posiposi/hona-movie/backend/internal/kernel"

// UserID はユーザーの識別子。kernel.ID を埋め込むことで、他集約の ID との
// 取り違えを型で防ぐ。
type UserID struct {
	kernel.ID
}

// NewUserID は新しいユーザー識別子を採番する。
func NewUserID() UserID {
	return UserID{ID: kernel.NewID()}
}

// ParseUserID は外部から受け取った文字列を UserID に変換する。ULID として
// 解釈できない値はドメインエラーとして拒否する。
func ParseUserID(value string) (UserID, error) {
	id, err := kernel.ParseID(value)
	if err != nil {
		return UserID{}, err
	}
	return UserID{ID: id}, nil
}
