package model

import "github.com/posiposi/hona-movie/backend/internal/kernel"

// UserID が kernel.ID をそのまま使わず埋め込むのは、他集約の ID との取り違えを
// 型で防ぐため。
type UserID struct {
	kernel.ID
}

func NewUserID() UserID {
	return UserID{ID: kernel.NewID()}
}

func ParseUserID(value string) (UserID, error) {
	id, err := kernel.ParseID(value)
	if err != nil {
		return UserID{}, err
	}
	return UserID{ID: id}, nil
}
