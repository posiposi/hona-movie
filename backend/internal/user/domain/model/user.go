package model

import "time"

type User struct {
	id        UserID
	name      UserName
	createdAt time.Time
	updatedAt time.Time
}

// CreateUser が createdAt / updatedAt をゼロ値のままにするのは、値が永続化時に
// 決まるため。
func CreateUser(name UserName) User {
	return User{
		id:   NewUserID(),
		name: name,
	}
}

// ReconstructUser がバリデーションを行わないのは、永続化済みの値は書き込み時に
// 検証を通っているため。
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

func (u User) Rename(name UserName) User {
	u.name = name
	return u
}
