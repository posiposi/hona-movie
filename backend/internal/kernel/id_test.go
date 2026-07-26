package kernel_test

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/posiposi/hona-movie/backend/internal/kernel"
)

const validULID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"

func TestNewID(t *testing.T) {
	t.Run("呼び出しごとに異なる値が生成される", func(t *testing.T) {
		first := kernel.NewID()
		second := kernel.NewID()

		if first.Equals(second) {
			t.Errorf("NewID().Equals(NewID()) = %v, want %v", true, false)
		}
	})

	t.Run("生成した値はULIDとして再パースできる", func(t *testing.T) {
		id := kernel.NewID()

		parsed, err := kernel.ParseID(id.String())
		if err != nil {
			t.Fatalf("ParseID(%v) error = %v, want nil", id.String(), err)
		}
		if !parsed.Equals(id) {
			t.Errorf("ParseID(%v).Equals(%v) = %v, want %v", id.String(), id, false, true)
		}
	})

	t.Run("連続生成した値が辞書順で単調増加する", func(t *testing.T) {
		const count = 1000
		prev := kernel.NewID().String()
		for i := 1; i < count; i++ {
			current := kernel.NewID().String()
			if current <= prev {
				t.Errorf("NewID().String() = %v, want greater than %v", current, prev)
				return
			}
			prev = current
		}
	})

	t.Run("並行に生成しても値が重複しない", func(t *testing.T) {
		const goroutines = 50
		const perGoroutine = 100

		var wg sync.WaitGroup
		results := make([]string, goroutines*perGoroutine)
		for g := range goroutines {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				for i := range perGoroutine {
					results[g*perGoroutine+i] = kernel.NewID().String()
				}
			}(g)
		}
		wg.Wait()

		seen := make(map[string]struct{}, len(results))
		for _, v := range results {
			if _, ok := seen[v]; ok {
				t.Errorf("NewID().String() = %v, want unique value", v)
				return
			}
			seen[v] = struct{}{}
		}
	})
}

func TestParseID(t *testing.T) {
	t.Run("妥当なULID文字列を保持できる", func(t *testing.T) {
		id, err := kernel.ParseID(validULID)
		if err != nil {
			t.Fatalf("ParseID(%v) error = %v, want nil", validULID, err)
		}

		if got := id.Value(); got != validULID {
			t.Errorf("ParseID(%v).Value() = %v, want %v", validULID, got, validULID)
		}
		if got := id.String(); got != validULID {
			t.Errorf("ParseID(%v).String() = %v, want %v", validULID, got, validULID)
		}
	})

	t.Run("前後の空白がtrimされて保存される", func(t *testing.T) {
		input := "  " + validULID + "  "

		id, err := kernel.ParseID(input)
		if err != nil {
			t.Fatalf("ParseID(%v) error = %v, want nil", input, err)
		}
		if got := id.Value(); got != validULID {
			t.Errorf("ParseID(%v).Value() = %v, want %v", input, got, validULID)
		}
	})

	t.Run("小文字は大文字に正規化して保持する", func(t *testing.T) {
		input := strings.ToLower(validULID)

		id, err := kernel.ParseID(input)
		if err != nil {
			t.Fatalf("ParseID(%v) error = %v, want nil", input, err)
		}
		if got := id.Value(); got != validULID {
			t.Errorf("ParseID(%v).Value() = %v, want %v", input, got, validULID)
		}
	})

	tests := []struct {
		name  string
		input string
	}{
		{"空文字はドメインエラーを返す", ""},
		{"空白のみはドメインエラーを返す", "   "},
		{"26文字未満はドメインエラーを返す", "01ARZ3NDEKTSV4RRFFQ69G5FA"},
		{"26文字を超える場合はドメインエラーを返す", "01ARZ3NDEKTSV4RRFFQ69G5FAVX"},
		{"Crockford Base32に含まれない文字はドメインエラーを返す", "01ARZ3NDEKTSV4RRFFQ69G5FAI"},
		{"上限を超えるULIDはドメインエラーを返す", "8ZZZZZZZZZZZZZZZZZZZZZZZZZ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kernel.ParseID(tt.input)
			if err == nil {
				t.Fatalf("ParseID(%v) = %v, want error", tt.input, got)
			}

			var domainErr *kernel.Error
			if !errors.As(err, &domainErr) {
				t.Fatalf("errors.As(%v, &domainErr) = false, want true", err)
			}
			if domainErr.Code == "" {
				t.Errorf("ParseID(%v) error Code = %v, want non-empty", tt.input, domainErr.Code)
			}
		})
	}
}

func TestID_Equals(t *testing.T) {
	t.Run("同じ値のID同士はtrueを返す", func(t *testing.T) {
		a, err := kernel.ParseID(validULID)
		if err != nil {
			t.Fatalf("ParseID(%v) error = %v, want nil", validULID, err)
		}
		b, err := kernel.ParseID(validULID)
		if err != nil {
			t.Fatalf("ParseID(%v) error = %v, want nil", validULID, err)
		}

		if !a.Equals(b) {
			t.Errorf("%v.Equals(%v) = %v, want %v", a, b, false, true)
		}
	})

	t.Run("異なる値のID同士はfalseを返す", func(t *testing.T) {
		a := kernel.NewID()
		b := kernel.NewID()

		if a.Equals(b) {
			t.Errorf("%v.Equals(%v) = %v, want %v", a, b, true, false)
		}
	})
}

func TestID_IsZero(t *testing.T) {
	t.Run("ゼロ値のIDはtrueを返す", func(t *testing.T) {
		var id kernel.ID

		if !id.IsZero() {
			t.Errorf("ID{}.IsZero() = %v, want %v", false, true)
		}
	})

	t.Run("生成したIDはfalseを返す", func(t *testing.T) {
		id := kernel.NewID()

		if id.IsZero() {
			t.Errorf("NewID().IsZero() = %v, want %v", true, false)
		}
	})
}
