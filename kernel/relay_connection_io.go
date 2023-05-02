package kernel

// Stellt eine Verbindungspaar dar
type _connection_io_pair struct {
	_relay *Relay
	_conn  []RelayConnection
}

// Diese Funktion gibt an ob es eine Aktive verbindung für diesen Relay gibt
func (obj *_connection_io_pair) HasActiveConnections() bool {
	return false
}

// Diese Funktion gibt an ob für diese Relay Verbindung bereits die Routen Initalisiert wurden
func (obj *_connection_io_pair) RoutestInited() bool {
	return false
}

// Fügt dem Paar eine Verbindung hinzu
func (obj *_connection_io_pair) add_connection(conn RelayConnection) error {
	return nil
}

// Erstellt ein neues Connection IO Pair
func createNewConnectionIoPair(relay *Relay) *_connection_io_pair {
	return &_connection_io_pair{_relay: relay}
}
