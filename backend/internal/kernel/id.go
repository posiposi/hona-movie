package kernel

import (
	"crypto/rand"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
)

const ErrCodeInvalidID = "INVALID_ID"

// ulid.Monotonic は並行安全ではないため、エントロピー源の読み出しは
// このミューテックスで保護する。
var (
	entropyMu     sync.Mutex
	monotonicRand = ulid.Monotonic(rand.Reader, 0)
)

// ID に ULID を採用しているのは、生成時刻順にソートできるようにするため。
type ID struct {
	value string
}

// NewID が返す識別子は、同一ミリ秒内に生成されたものどうしでも生成順を保つ。
func NewID() ID {
	entropyMu.Lock()
	defer entropyMu.Unlock()
	return ID{value: ulid.MustNew(ulid.Now(), monotonicRand).String()}
}

func ParseID(value string) (ID, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return ID{}, NewError(ErrCodeInvalidID, "id must not be empty")
	}
	// ulid.Parse は速度優先で文字検証を省くため、不正な文字列を黙って別の値に
	// デコードしてしまう。厳密版を使って弾く。
	parsed, err := ulid.ParseStrict(normalized)
	if err != nil {
		return ID{}, NewError(ErrCodeInvalidID, "id must be a valid ULID")
	}
	return ID{value: parsed.String()}, nil
}

func (i ID) Value() string {
	return i.value
}

func (i ID) String() string {
	return i.value
}

func (i ID) Equals(other ID) bool {
	return i.value == other.value
}

func (i ID) IsZero() bool {
	return i.value == ""
}
