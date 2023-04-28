package kernel

import (
	"fmt"
	"sync"
)

type _connection_io_pair struct {
	_relay *Relay
	_conn  []*RelayConnection
}

type ConnectionManager struct {
	_lock       *sync.Mutex
	_connection []*_connection_io_pair
}

func (obj *ConnectionManager) RegisterNewRelayConnection(relay *Relay, conn RelayConnection) error {
	// Es wird geprüft ob es bereits ein Relay mit diesem Öffentlichen Schlüssel gibt
	for i := range obj._connection {
		if obj._connection[i]._relay._public_key == relay._public_key {
			// Die Verbindung wird dem Relay zugeordnet
			obj._connection[i]._conn = append(obj._connection[i]._conn, &conn)
			return nil
		}
	}

	// Es wurde kein Relay gefunden, es wird ein neuer Haupteintrag erzeugt
	obj._connection = append(obj._connection, &_connection_io_pair{_relay: relay, _conn: []*RelayConnection{&conn}})

	// Der Vorgang wurde ohne Fehler erfolgreich druchgeführt
	return nil
}

func (obj *ConnectionManager) RemoveRelayConnection(relay *Relay, conn RelayConnection) error {
	// Es wird geprüft um welches Relay es sich handelt, dieser wird entfernt
	var found_id int64
	var found bool
	for i := range obj._connection {
		if obj._connection[i]._relay._public_key == relay._public_key {
			found_id = int64(i)
			found = true
			break
		}
	}

	// Sollte der Eintrag nicht gefunden werden, wird der Vorgang abgebrochen
	if !found {
		return fmt.Errorf("relay not found")
	}

	// Der Eintrag wird entfernt

	// Der Vorgang wurde ohne fehler erfolgreich durchgeführt
	return nil
}

func (obj *ConnectionManager) RelayIsConnected(relay *Relay) bool {
	for i := range obj._connection {
		if obj._connection[i]._relay._public_key == relay._public_key {
			return true
		}
	}
	return false
}

func newConnectionManager() ConnectionManager {
	return ConnectionManager{_connection: make([]*_connection_io_pair, 0), _lock: new(sync.Mutex)}
}
