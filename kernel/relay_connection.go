package kernel

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"
)

// Stellt den Verbindungsmanager dar
type ConnectionManager struct {
	_lock       *sync.Mutex
	_connection []*connection_io_pair
}

// Gibt an ob es eine Aktive Verbindung gibt
func (obj *ConnectionManager) HasActiveConnections() bool {
	obj._lock.Lock()
	for i := range obj._connection {
		if obj._connection[i].HasActiveConnections() {
			obj._lock.Unlock()
			return true
		}
	}
	obj._lock.Unlock()
	return false
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
			log.Println("ConnectionManager: relay connection added. relay =", relay._hexed_id, "connection =", conn.GetObjectId())
			obj._connection[i].add_connection(conn)
			obj._lock.Unlock()
			return nil
		}
	}

	// Es wird eine neuer RelayIO erstellt
	relay_io_object := createNewConnectionIoPair(relay)

	// Dem Relay wird eine Verbindung zugeordnet
	err := relay_io_object.add_connection(conn)
	if err != nil {
		obj._lock.Unlock()
		return fmt.Errorf("ConnectionManager:RegisterNewRelayConnection: " + err.Error())
	}

	// Es wurde kein Relay gefunden, es wird ein neuer Haupteintrag erzeugt
	obj._connection = append(obj._connection, relay_io_object)

	// Der Thradlock wird freigegeben
	obj._lock.Unlock()

	// Log
	log.Println("ConnectionManager: new Relay added. relay =", relay.GetPublicKeyHexString(), "connection =", conn.GetObjectId())

	// Der Vorgang wurde ohne Fehler erfolgreich druchgeführt
	return nil
}

// Entfernt eine Verbindung
func (obj *ConnectionManager) RemoveConnectionFromRelay(conn RelayConnection) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob es bereits ein Relay mit diesem Öffentlichen Schlüssel gibt
	var has_found_entry bool
	for i := range obj._connection {
		if obj._connection[i]._is_used_conenction(conn) {
			if err := obj._connection[i].remove_connection(conn); err != nil {
				obj._lock.Unlock()
				return fmt.Errorf("ConnectionManager: RemoveConnectionFromRelay: 1: " + err.Error())
			}
			has_found_entry = true
			break
		}
	}

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Sollte kein Eintrag gefunden wurden sein, wird der Vorgang abgebrochen
	if !has_found_entry {
		log.Println("ConnectionManager: can't remove unkown connection. connection =", conn.GetObjectId())
		return nil
	}

	// Der Vorgang wurde ohne fehler erfolgreich durchgeführt
	return nil
}

// Gibt an ob der Relay Verbunden ist
func (obj *ConnectionManager) RelayIsConnected(relay *Relay) bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob es eine Aktive Verbindung für das Relay gibt
	for i := range obj._connection {
		if obj._connection[i]._relay._hexed_id == relay._hexed_id {
			obj._lock.Unlock()
			return true
		}
	}

	// Der Threadlock wird freigegben
	obj._lock.Unlock()

	// Es wurde keine Aktive Realy Verbinding gefunden
	return false
}

// Ruft ein Relay anhand einer Verbindung ab
func (obj *ConnectionManager) GetRelayByConnection(conn RelayConnection) (*Relay, bool, error) {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob es eine Verbindung mit dieser ID gibt
	var has_found bool
	var frelay *Relay
	for i := range obj._connection {
		for x := range obj._connection[i]._conn {
			if obj._connection[i]._conn[x].GetObjectId() == conn.GetObjectId() {
				frelay = obj._connection[i]._relay
				has_found = true
				break
			}
		}
		if has_found {
			break
		}
	}

	// Sollte kein Relay gefunden wurden sein, wird der Vorgang abgebrochen
	if !has_found {
		obj._lock.Unlock()
		return nil, false, nil
	}

	// Der Threadlock wird feigegeben
	obj._lock.Unlock()

	// Der Vorgang wurde ohne fehler abgeschlossen
	return frelay, true, nil
}

// Wird verwendet um alle Aktiven Routen für ein Relay zu Initalisieren
func (obj *ConnectionManager) InitRoutesForRelay(relay *Relay, routes []*RouteEntry) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Das Relay wird herausgesucht
	for i := range obj._connection {
		if obj._connection[i]._relay._hexed_id == relay._hexed_id {
			// Log
			log.Println("ConnectionManager: try initing routes for relay. relay =", relay._hexed_id)

			// Es wird geprüft ob es eine Aktive verbindung für dieses Relay gibt
			if !obj._connection[i].HasActiveConnections() {
				obj._lock.Unlock()
				return fmt.Errorf("ConnectionManager: relay has no active connections. relay = " + relay.GetHexId())
			}

			// Das Relay wird aktiviert
			if err := obj._connection[i]._signal_activated(); err != nil {
				obj._lock.Unlock()
				return fmt.Errorf("InitRoutesForRelay: 1:" + err.Error())
			}

			// Die Routen werden dem Relay zugeordnet
			if err := obj._connection[i]._add_route_entrys(routes); err != nil {
				obj._lock.Unlock()
				return fmt.Errorf("InitRoutesForRelay: 2: " + err.Error())
			}

			// Die Routen würd das Relay wurden erfolgreich bereitgestellt
			obj._lock.Unlock()
			return nil
		}
	}

	// Der Threadlock wird freigeben
	obj._lock.Unlock()

	// Das Relay wird herausgesucht
	return fmt.Errorf("ConnectionManager: InitRoutesForRelay: no relay found")
}

// Wird vom Kernel verwendet alle Verbindungen zu schließen
func (obj *ConnectionManager) ShutdownByKernel() {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird allen Relays Signalisiert dass sie all ihre Verbindungen schließen sollen
	relist := obj._connection

	// Der Threadlock wird freigeben
	obj._lock.Unlock()

	// Die Verbindungen werden geschlossen
	for i := range relist {
		relist[i].kernel_shutdown()
	}

	// Es wird gewartet bis alle Verbindungen geschlossen wurden
	for obj.HasActiveConnections() {
		time.Sleep(1 * time.Millisecond)
	}
}

// Erstellt einen neuen Verbindungs Manager
func newConnectionManager() ConnectionManager {
	return ConnectionManager{_connection: make([]*connection_io_pair, 0), _lock: new(sync.Mutex)}
}
