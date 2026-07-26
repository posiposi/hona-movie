package model_test

import (
	"errors"
	"testing"

	"github.com/posiposi/hona-movie/backend/internal/kernel"
	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
)

func TestNewUserID(t *testing.T) {
	t.Run("生成したUserIDは空でない", func(t *testing.T) {
		id := model.NewUserID()

		if id.Value() == "" {
			t.Errorf("NewUserID().Value() = %v, want non-empty", id.Value())
		}
	})

	t.Run("生成するたびに異なる値になる", func(t *testing.T) {
		a := model.NewUserID()
		b := model.NewUserID()

		if a.Equals(b.ID) {
			t.Errorf("NewUserID() = %v, want different from %v", b.Value(), a.Value())
		}
	})
}

func TestParseUserID(t *testing.T) {
	t.Run("有効なULIDからUserIDを生成できる", func(t *testing.T) {
		want := model.NewUserID().Value()

		got, err := model.ParseUserID(want)
		if err != nil {
			t.Fatalf("ParseUserID(%v) returned unexpected error: %v", want, err)
		}

		if got.Value() != want {
			t.Errorf("ParseUserID(%v).Value() = %v, want %v", want, got.Value(), want)
		}
	})

	t.Run("不正な値の場合はドメインエラーを返す", func(t *testing.T) {
		tests := []struct {
			name  string
			value string
		}{
			{name: "空文字", value: ""},
			{name: "空白のみ", value: "   "},
			{name: "ULIDでない文字列", value: "not-a-ulid"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := model.ParseUserID(tt.value)
				if err == nil {
					t.Fatalf("ParseUserID(%q) = nil, want error", tt.value)
				}

				var domainErr *kernel.Error
				if !errors.As(err, &domainErr) {
					t.Fatalf("ParseUserID(%q) error = %v, want *kernel.Error", tt.value, err)
				}
				if domainErr.Code != kernel.ErrCodeInvalidID {
					t.Errorf("ParseUserID(%q) error code = %v, want %v", tt.value, domainErr.Code, kernel.ErrCodeInvalidID)
				}
			})
		}
	})
}
