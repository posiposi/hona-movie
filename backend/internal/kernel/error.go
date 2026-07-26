// Package kernel は全モジュールが共有するドメインの基盤型を提供する。
// 業務ロジックはここに置かない。
package kernel

// Error はドメインエラーを表す。呼び出し側がメッセージ文字列をパースせずに
// 分岐できるよう、安定したコードを持たせる。
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
