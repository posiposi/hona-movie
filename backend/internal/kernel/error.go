// Package kernel は全モジュールが共有するドメインの基盤型を置く場所であり、
// 業務ロジックは持ち込まない。
package kernel

// Error が Code を持つのは、呼び出し側がメッセージ文字列をパースせずに
// 分岐できるようにするため。
type Error struct {
	Code    string
	Message string
}

func NewError(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func (e *Error) Error() string {
	return e.Message
}
