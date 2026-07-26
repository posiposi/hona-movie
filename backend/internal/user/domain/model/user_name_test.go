package model_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/posiposi/hona-movie/backend/internal/kernel"
	"github.com/posiposi/hona-movie/backend/internal/user/domain/model"
)

func TestNewUserName(t *testing.T) {
	t.Run("有効な値でUserNameを生成できる", func(t *testing.T) {
		tests := []struct {
			name  string
			value string
			want  string
		}{
			{name: "英字", value: "posipos", want: "posipos"},
			{name: "日本語", value: "杉山大地", want: "杉山大地"},
			{name: "前後の空白はトリムされる", value: "  posipos  ", want: "posipos"},
			{name: "1文字", value: "a", want: "a"},
			{name: "255文字ちょうど", value: strings.Repeat("a", 255), want: strings.Repeat("a", 255)},
			{name: "マルチバイト255文字ちょうど", value: strings.Repeat("あ", 255), want: strings.Repeat("あ", 255)},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := model.NewUserName(tt.value)
				if err != nil {
					t.Fatalf("NewUserName(%q) returned unexpected error: %v", tt.value, err)
				}

				if got.Value() != tt.want {
					t.Errorf("NewUserName(%q).Value() = %v, want %v", tt.value, got.Value(), tt.want)
				}
			})
		}
	})

	t.Run("不正な値の場合はドメインエラーを返す", func(t *testing.T) {
		tests := []struct {
			name  string
			value string
		}{
			{name: "空文字", value: ""},
			{name: "空白のみ", value: "   "},
			{name: "タブと改行のみ", value: "\t\n"},
			{name: "256文字", value: strings.Repeat("a", 256)},
			{name: "マルチバイト256文字", value: strings.Repeat("あ", 256)},
			{name: "トリム後に256文字", value: "  " + strings.Repeat("a", 256) + "  "},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := model.NewUserName(tt.value)
				if err == nil {
					t.Fatalf("NewUserName(%q) = nil, want error", tt.value)
				}

				var domainErr *kernel.Error
				if !errors.As(err, &domainErr) {
					t.Fatalf("NewUserName(%q) error = %v, want *kernel.Error", tt.value, err)
				}
				if domainErr.Code != model.ErrCodeInvalidUserName {
					t.Errorf("NewUserName(%q) error code = %v, want %v", tt.value, domainErr.Code, model.ErrCodeInvalidUserName)
				}
			})
		}
	})
}

func TestUserName_Equals(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		want  bool
	}{
		{name: "同じ値はtrue", left: "posipos", right: "posipos", want: true},
		{name: "トリム後に同じ値はtrue", left: "posipos", right: "  posipos  ", want: true},
		{name: "異なる値はfalse", left: "posipos", right: "daichi", want: false},
		{name: "大文字小文字が異なる場合はfalse", left: "posipos", right: "POSIPOS", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left, err := model.NewUserName(tt.left)
			if err != nil {
				t.Fatalf("NewUserName(%q) returned unexpected error: %v", tt.left, err)
			}
			right, err := model.NewUserName(tt.right)
			if err != nil {
				t.Fatalf("NewUserName(%q) returned unexpected error: %v", tt.right, err)
			}

			if got := left.Equals(right); got != tt.want {
				t.Errorf("UserName.Equals(%v) = %v, want %v", tt.right, got, tt.want)
			}
		})
	}
}
