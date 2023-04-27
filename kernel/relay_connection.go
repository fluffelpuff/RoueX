package kernel

type _connection_io_pair struct {
}

type ConnectionManager struct {
	_connection []*_connection_io_pair
}

func newConnectionManager() ConnectionManager {
	return ConnectionManager{}
}
