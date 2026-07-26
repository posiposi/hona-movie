//go:build integration

package mysql_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
	"github.com/posiposi/hona-movie/backend/internal/user/domain/repository"
	usermysql "github.com/posiposi/hona-movie/backend/internal/user/infrastructure/mysql"
)

type userRow struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func selectUserRow(t *testing.T, id model.UserID) (userRow, bool) {
	t.Helper()
	var rows []userRow
	err := testDB.Raw("SELECT id, name, created_at, updated_at FROM users WHERE id = ?", id.Value()).Scan(&rows).Error
	if err != nil {
		t.Fatalf("select user(%v) returned error: %v", id.Value(), err)
	}
	if len(rows) == 0 {
		return userRow{}, false
	}
	return rows[0], true
}

func TestUserCommandRepositorySave(t *testing.T) {
	repo := usermysql.NewUserCommandRepository(testDB)
	ctx := context.Background()

	t.Run("新規ユーザーを登録できる", func(t *testing.T) {
		user := model.CreateUser(newTestUserName(t, "save-create"))
		cleanupUsers(t, user.ID())

		if err := repo.Save(ctx, user); err != nil {
			t.Fatalf("Save(%v) returned error: %v", user.ID().Value(), err)
		}

		row, ok := selectUserRow(t, user.ID())
		if !ok {
			t.Fatalf("Save(%v) stored no row, want one row", user.ID().Value())
		}
		if row.Name != user.Name().Value() {
			t.Errorf("Save(%v).name = %v, want %v", user.ID().Value(), row.Name, user.Name().Value())
		}
		if row.CreatedAt.IsZero() {
			t.Errorf("Save(%v).created_at = %v, want non-zero", user.ID().Value(), row.CreatedAt)
		}
		if row.UpdatedAt.IsZero() {
			t.Errorf("Save(%v).updated_at = %v, want non-zero", user.ID().Value(), row.UpdatedAt)
		}
	})

	t.Run("既存ユーザーを更新でき、created_atは維持される", func(t *testing.T) {
		id := model.NewUserID()
		name := newTestUserName(t, "save-update")
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		insertTestUser(t, id, name, createdAt)

		renamed := model.ReconstructUser(id, name, createdAt, createdAt).Rename(newTestUserName(t, "save-updated"))
		if err := repo.Save(ctx, renamed); err != nil {
			t.Fatalf("Save(%v) returned error: %v", id.Value(), err)
		}

		row, ok := selectUserRow(t, id)
		if !ok {
			t.Fatalf("Save(%v) stored no row, want one row", id.Value())
		}
		if row.Name != renamed.Name().Value() {
			t.Errorf("Save(%v).name = %v, want %v", id.Value(), row.Name, renamed.Name().Value())
		}
		if got := row.CreatedAt.UTC(); !got.Equal(createdAt) {
			t.Errorf("Save(%v).created_at = %v, want %v", id.Value(), got, createdAt)
		}
		if got := row.UpdatedAt.UTC(); !got.After(createdAt) {
			t.Errorf("Save(%v).updated_at = %v, want after %v", id.Value(), got, createdAt)
		}
	})
}

func TestUserCommandRepositoryDelete(t *testing.T) {
	repo := usermysql.NewUserCommandRepository(testDB)
	ctx := context.Background()

	t.Run("既存ユーザーを削除できる", func(t *testing.T) {
		id := model.NewUserID()
		insertTestUser(t, id, newTestUserName(t, "delete"), time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC))

		if err := repo.Delete(ctx, id); err != nil {
			t.Fatalf("Delete(%v) returned error: %v", id.Value(), err)
		}
		if _, ok := selectUserRow(t, id); ok {
			t.Errorf("Delete(%v) left the row, want it removed", id.Value())
		}
	})

	t.Run("存在しないIDではErrUserNotFoundを返す", func(t *testing.T) {
		id := model.NewUserID()

		if err := repo.Delete(ctx, id); !errors.Is(err, repository.ErrUserNotFound) {
			t.Errorf("Delete(%v) error = %v, want %v", id.Value(), err, repository.ErrUserNotFound)
		}
	})
}
