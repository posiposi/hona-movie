package model_test

import (
	"testing"
	"time"

	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
)

func newTestUserName(t *testing.T, value string) model.UserName {
	t.Helper()
	name, err := model.NewUserName(value)
	if err != nil {
		t.Fatalf("NewUserName(%q) returned unexpected error: %v", value, err)
	}
	return name
}

func TestCreateUser(t *testing.T) {
	t.Run("IDを採番しタイムスタンプはゼロ値のままユーザーを生成する", func(t *testing.T) {
		name := newTestUserName(t, "posipos")

		u := model.CreateUser(name)

		if u.ID().IsZero() {
			t.Errorf("CreateUser(%v).ID() = %v, want non-zero", name.Value(), u.ID().Value())
		}
		if !u.Name().Equals(name) {
			t.Errorf("CreateUser(%v).Name() = %v, want %v", name.Value(), u.Name().Value(), name.Value())
		}
		if !u.CreatedAt().IsZero() {
			t.Errorf("CreateUser(%v).CreatedAt() = %v, want zero value", name.Value(), u.CreatedAt())
		}
		if !u.UpdatedAt().IsZero() {
			t.Errorf("CreateUser(%v).UpdatedAt() = %v, want zero value", name.Value(), u.UpdatedAt())
		}
	})

	t.Run("生成するたびに異なるIDが採番される", func(t *testing.T) {
		name := newTestUserName(t, "posipos")

		a := model.CreateUser(name)
		b := model.CreateUser(name)

		if a.ID().Equals(b.ID().ID) {
			t.Errorf("CreateUser(%v).ID() = %v, want different from %v", name.Value(), b.ID().Value(), a.ID().Value())
		}
	})
}

func TestReconstructUser(t *testing.T) {
	t.Run("指定したすべての値でユーザーを復元する", func(t *testing.T) {
		id := model.NewUserID()
		name := newTestUserName(t, "posipos")
		createdAt := time.Date(2026, 7, 26, 1, 43, 29, 0, time.UTC)
		updatedAt := time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC)

		u := model.ReconstructUser(id, name, createdAt, updatedAt)

		if !u.ID().Equals(id.ID) {
			t.Errorf("ReconstructUser(%v).ID() = %v, want %v", id.Value(), u.ID().Value(), id.Value())
		}
		if !u.Name().Equals(name) {
			t.Errorf("ReconstructUser(%v).Name() = %v, want %v", id.Value(), u.Name().Value(), name.Value())
		}
		if !u.CreatedAt().Equal(createdAt) {
			t.Errorf("ReconstructUser(%v).CreatedAt() = %v, want %v", id.Value(), u.CreatedAt(), createdAt)
		}
		if !u.UpdatedAt().Equal(updatedAt) {
			t.Errorf("ReconstructUser(%v).UpdatedAt() = %v, want %v", id.Value(), u.UpdatedAt(), updatedAt)
		}
	})
}

func TestUser_Rename(t *testing.T) {
	t.Run("名前だけを差し替えた新しいUserを返す", func(t *testing.T) {
		id := model.NewUserID()
		createdAt := time.Date(2026, 7, 26, 1, 43, 29, 0, time.UTC)
		updatedAt := time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC)
		original := model.ReconstructUser(id, newTestUserName(t, "posipos"), createdAt, updatedAt)
		newName := newTestUserName(t, "daichi")

		renamed := original.Rename(newName)

		if !renamed.Name().Equals(newName) {
			t.Errorf("User.Rename(%v).Name() = %v, want %v", newName.Value(), renamed.Name().Value(), newName.Value())
		}
		if !renamed.ID().Equals(id.ID) {
			t.Errorf("User.Rename(%v).ID() = %v, want %v", newName.Value(), renamed.ID().Value(), id.Value())
		}
		if !renamed.CreatedAt().Equal(createdAt) {
			t.Errorf("User.Rename(%v).CreatedAt() = %v, want %v", newName.Value(), renamed.CreatedAt(), createdAt)
		}
		if !renamed.UpdatedAt().Equal(updatedAt) {
			t.Errorf("User.Rename(%v).UpdatedAt() = %v, want %v", newName.Value(), renamed.UpdatedAt(), updatedAt)
		}
	})

	t.Run("レシーバのUserは変更されない", func(t *testing.T) {
		original := model.ReconstructUser(model.NewUserID(), newTestUserName(t, "posipos"), time.Time{}, time.Time{})

		original.Rename(newTestUserName(t, "daichi"))

		if got := original.Name().Value(); got != "posipos" {
			t.Errorf("User.Name() after Rename = %v, want %v", got, "posipos")
		}
	})
}
