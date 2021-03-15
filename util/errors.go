package util

type TokenSizeError struct {
	Msg string
}

func (t *TokenSizeError) Error() string {
	return t.Msg
}
