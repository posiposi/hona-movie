//go:build integration

package mysql_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
	"github.com/posiposi/hona-movie/backend/internal/user/domain/repository"
	usermysql "github.com/posiposi/hona-movie/backend/internal/user/infrastructure/mysql"
)

func TestUserQueryRepositoryFindByID(t *testing.T) {
	repo := usermysql.NewUserQueryRepository(testDB)
	ctx := context.Background()

	t.Run("保存済みのユーザーを取得できる", func(t *testing.T) {
		id := model.NewUserID()
		name := newTestUserName(t, "find-by-id")
		at := time.Date(2026, 7, 26, 1, 2, 3, 0, time.UTC)
		insertTestUser(t, id, name, at)

		found, err := repo.FindByID(ctx, id)
		if err != nil {
			t.Fatalf("FindByID(%v) returned error: %v", id.Value(), err)
		}
		if found == nil {
			t.Fatalf("FindByID(%v) = nil, want non-nil", id.Value())
		}
		if got := found.ID().Value(); got != id.Value() {
			t.Errorf("FindByID(%v).ID() = %v, want %v", id.Value(), got, id.Value())
		}
		if got := found.Name().Value(); got != name.Value() {
			t.Errorf("FindByID(%v).Name() = %v, want %v", id.Value(), got, name.Value())
		}
		if got := found.CreatedAt().UTC(); !got.Equal(at) {
			t.Errorf("FindByID(%v).CreatedAt() = %v, want %v", id.Value(), got, at)
		}
		if got := found.UpdatedAt().UTC(); !got.Equal(at) {
			t.Errorf("FindByID(%v).UpdatedAt() = %v, want %v", id.Value(), got, at)
		}
	})

	t.Run("存在しないIDではErrUserNotFoundを返す", func(t *testing.T) {
		id := model.NewUserID()

		found, err := repo.FindByID(ctx, id)
		if !errors.Is(err, repository.ErrUserNotFound) {
			t.Errorf("FindByID(%v) error = %v, want %v", id.Value(), err, repository.ErrUserNotFound)
		}
		if found != nil {
			t.Errorf("FindByID(%v) = %v, want nil", id.Value(), found)
		}
	})
}

func TestUserQueryRepositoryFindAll(t *testing.T) {
	repo := usermysql.NewUserQueryRepository(testDB)
	ctx := context.Background()

	t.Run("保存済みのユーザーがすべて含まれる", func(t *testing.T) {
		at := time.Date(2026, 7, 26, 4, 5, 6, 0, time.UTC)
		firstID := model.NewUserID()
		firstName := newTestUserName(t, "find-all-1")
		insertTestUser(t, firstID, firstName, at)
		secondID := model.NewUserID()
		secondName := newTestUserName(t, "find-all-2")
		insertTestUser(t, secondID, secondName, at)

		users, err := repo.FindAll(ctx)
		if err != nil {
			t.Fatalf("FindAll() returned error: %v", err)
		}

		found := make(map[string]string)
		for _, u := range users {
			if strings.HasPrefix(u.Name().Value(), testDataPrefix) {
				found[u.ID().Value()] = u.Name().Value()
			}
		}
		if got := found[firstID.Value()]; got != firstName.Value() {
			t.Errorf("FindAll()[%v] = %v, want %v", firstID.Value(), got, firstName.Value())
		}
		if got := found[secondID.Value()]; got != secondName.Value() {
			t.Errorf("FindAll()[%v] = %v, want %v", secondID.Value(), got, secondName.Value())
		}
	})
}
