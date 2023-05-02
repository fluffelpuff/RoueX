package kernel

import (
	"bytes"
	"fmt"
	"log"
	"sync"
)

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

// Stellt den Verbindungsmanager dar
type ConnectionManager struct {
	_lock       *sync.Mutex
	_connection []*_connection_io_pair
}

// Fügt eine neue AKtive Verbindung zum Manager hinzu
func (obj *ConnectionManager) RegisterNewRelayConnection(relay *Relay, conn RelayConnection) error {
	// Es wird geprüft ob ein Relay vorhanden ist, wenn nicht wird ein Fehler produziert
	if relay == nil {
		return fmt.Errorf("RegisterNewRelayConnection: you need a relay object")
	}

	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob es bereits ein Relay mit diesem Öffentlichen Schlüssel gibt
	for i := range obj._connection {
		if bytes.Equal(obj._connection[i]._relay._public_key.SerializeCompressed(), relay._public_key.SerializeCompressed()) {
			// Die Verbindung wird dem Relay zugeordnet
			log.Println("Relay connection added", relay.GetPublicKeyHexString())
			obj._connection[i]._conn = append(obj._connection[i]._conn, conn)
			return nil
		}
	}

	// Es wurde kein Relay gefunden, es wird ein neuer Haupteintrag erzeugt
	obj._connection = append(obj._connection, &_connection_io_pair{_relay: relay, _conn: []RelayConnection{conn}})

	// Der Thradlock wird freigegeben
	obj._lock.Unlock()

	// Log
	log.Println("ConnectionManager: new Relay added. relay =", relay.GetPublicKeyHexString(), "connection =", conn.GetObjectId())

	// Der Vorgang wurde ohne Fehler erfolgreich druchgeführt
	return nil
}

// Entfernt eine Verbindung
func (obj *ConnectionManager) RemoveRelayConnection(conn RelayConnection) error {
	// Log
	log.Println("ConnectionManager: connection removed. connection =", conn.GetObjectId())

	// Der Vorgang wurde ohne fehler erfolgreich durchgeführt
	return nil
}

// Gibt an ob der Relay Verbunden ist
func (obj *ConnectionManager) RelayIsConnected(relay *Relay) bool {
	for i := range obj._connection {
		if obj._connection[i]._relay._public_key == relay._public_key {
			return true
		}
	}
	return false
}

// Ruft ein Relay anhand einer Verbindung ab
func (obj *ConnectionManager) GetRelayByConnection(conn RelayConnection) (*Relay, bool, error) {
	// Es wird geprüft ob der Client eine Verbindung besitzt
	if !conn.IsConnected() {
		return nil, false, fmt.Errorf("ConnectionManager:GetRelayByConnection: connection isn connected")
	}

	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob es eine Verbindung mit dieser ID gibt
	var frelay *Relay
	for i := range obj._connection {
		for x := range obj._connection[i]._conn {
			if obj._connection[i]._conn[x].GetObjectId() == conn.GetObjectId() {
				frelay = obj._connection[i]._relay
				break
			}
		}
	}

	// Sollte kein Relay gefunden wurden sein, wird der Vorgang abgebrochen
	if frelay != nil {
		obj._lock.Unlock()
		return nil, false, nil
	}

	// Der Threadlock wird feigegeben
	obj._lock.Unlock()

	// Der Vorgang wurde ohne fehler abgeschlossen
	return nil, true, nil
}

// Gibt an ob für diesen Relay bereits die Routen Initalisiert wurden und ob eine Aktive Verbidndung für den Relay vorhanden ist
func (obj *ConnectionManager) RelayAvailAndRoutesInited(relay *Relay) (bool, bool) {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Der Relay wird herausgefiltert
	for i := range obj._connection {
		// Es wird nach dem Relay gesucht
		if obj._connection[i]._relay._hexed_id == relay._hexed_id {
			// Es wird geprüft ob der Relay mindestens eine Aktive verbindung hat
			if obj._connection[i].HasActiveConnections() {
				// Es wird geprüft ob die Routen bereits Initalisiert wurden
				if obj._connection[i].RoutestInited() {
					obj._lock.Unlock()
					return true, true
				} else {
					obj._lock.Unlock()
					return true, false
				}
			}
		}
	}

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Es wurde kein Aktiver Relay mit Initalisierten Routen gefunden
	return false, false
}

// Wird verwendet um alle Aktiven Routen für ein Relay zu Initalisieren
func (obj *ConnectionManager) InitRoutesForRelay(relay *Relay, routes []*RouteEntry) error {
	return nil
}

// Erstellt einen neuen Verbindungs Manager
func newConnectionManager() ConnectionManager {
	return ConnectionManager{_connection: make([]*_connection_io_pair, 0), _lock: new(sync.Mutex)}
}
