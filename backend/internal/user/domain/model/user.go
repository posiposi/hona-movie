package model

import "time"

// User はユーザー集約のルートエンティティ。
type User struct {
	id        UserID
	name      UserName
	createdAt time.Time
	updatedAt time.Time
}

// CreateUser は新規ユーザーを生成する。createdAt / updatedAt は永続化時に DB の
// 既定値で決まるため、ここではゼロ値のままにする。
func CreateUser(name UserName) User {
	return User{
		id:   NewUserID(),
		name: name,
	}
}

// ReconstructUser は永続化済みの値から User を復元する。DB からの復元専用で
// あり、ID の採番もバリデーションも行わない。
func ReconstructUser(id UserID, name UserName, createdAt, updatedAt time.Time) User {
	return User{
		id:        id,
		name:      name,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (u User) ID() UserID {
	return u.id
}

func (u User) Name() UserName {
	return u.name
}

func (u User) CreatedAt() time.Time {
	return u.createdAt
}

func (u User) UpdatedAt() time.Time {
	return u.updatedAt
}

// Rename は名前を差し替えた新しい User を返す。レシーバは変更しない。
func (u User) Rename(name UserName) User {
	u.name = name
	return u
}
