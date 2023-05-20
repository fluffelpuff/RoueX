package rerror

type IOStateError struct{}

func (m *IOStateError) Error() string {
	return "boom"
}

func NewIOStateError() IOStateError {
	return IOStateError{}
}
