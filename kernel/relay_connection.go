package kernel

import (
	"bytes"
	"fmt"
	"log"
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
	// Es wird geprüft ob ein Relay vorhanden ist, wenn nicht wird ein Fehler produziert
	if relay == nil {
		return fmt.Errorf("RegisterNewRelayConnection: you need a relay object")
	}

	// Es wird geprüft ob es bereits ein Relay mit diesem Öffentlichen Schlüssel gibt
	for i := range obj._connection {
		if bytes.Equal(obj._connection[i]._relay._public_key.SerializeCompressed(), relay._public_key.SerializeCompressed()) {
			// Die Verbindung wird dem Relay zugeordnet
			log.Println("Relay connection added", relay.GetPublicKeyHexString())
			obj._connection[i]._conn = append(obj._connection[i]._conn, &conn)
			return nil
		}
	}

	// Es wurde kein Relay gefunden, es wird ein neuer Haupteintrag erzeugt
	obj._connection = append(obj._connection, &_connection_io_pair{_relay: relay, _conn: []*RelayConnection{&conn}})

	// Log
	log.Println("New Relay added", relay.GetPublicKeyHexString())

	// Der Vorgang wurde ohne Fehler erfolgreich druchgeführt
	return nil
}

func (obj *ConnectionManager) RemoveRelayConnection(conn RelayConnection) error {
	// Es werden alle Registrierten Paare nach dieser Verbindung abgesucht

	// Log
	log.Println("Relay removed", conn.GetObjectId())

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
