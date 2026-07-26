package kernel_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/posiposi/hona-movie/backend/internal/kernel"
)

func TestNewError(t *testing.T) {
	t.Run("生成したErrorがcodeとmessageを保持する", func(t *testing.T) {
		err := kernel.NewError("INVALID_ID", "IDの形式が不正です")

		if err.Code != "INVALID_ID" {
			t.Errorf("NewError(%v, %v).Code = %v, want %v", "INVALID_ID", "IDの形式が不正です", err.Code, "INVALID_ID")
		}
		if err.Message != "IDの形式が不正です" {
			t.Errorf("NewError(%v, %v).Message = %v, want %v", "INVALID_ID", "IDの形式が不正です", err.Message, "IDの形式が不正です")
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Run("Errorがmessage文字列を返す", func(t *testing.T) {
		err := kernel.NewError("MOVIE_NOT_FOUND", "映画が見つかりません")

		if got := err.Error(); got != "映画が見つかりません" {
			t.Errorf("Error.Error() = %v, want %v", got, "映画が見つかりません")
		}
	})
}

func TestError_ErrorsAs(t *testing.T) {
	t.Run("ラップされたerrorからerrors.Asで取り出してCodeへアクセスできる", func(t *testing.T) {
		var err error = kernel.NewError("ALREADY_EXISTS", "既に登録されています")
		wrapped := fmt.Errorf("usecase failed: %w", err)

		var domainErr *kernel.Error
		if !errors.As(wrapped, &domainErr) {
			t.Fatalf("errors.As(%v, &domainErr) = false, want true", wrapped)
		}
		if domainErr.Code != "ALREADY_EXISTS" {
			t.Errorf("domainErr.Code = %v, want %v", domainErr.Code, "ALREADY_EXISTS")
		}
	})
}
