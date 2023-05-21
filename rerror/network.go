package rerror

type IOStateError struct {
	_msg string
}

func (m *IOStateError) Error() string {
	return m._msg
}

func NewIOStateError(msg string) *IOStateError {
	return &IOStateError{msg}
}
