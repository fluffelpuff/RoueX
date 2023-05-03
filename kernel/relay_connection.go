package kernel

import (
	"bytes"
	"fmt"
	"log"
	"sync"
)

// Stellt den Verbindungsmanager dar
type ConnectionManager struct {
	_lock       *sync.Mutex
	_connection []*connection_io_pair
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
func (obj *ConnectionManager) RemoveRelayConnection(conn RelayConnection) error {
	// Log
	log.Println("ConnectionManager: connection removed. connection =", conn.GetObjectId())

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
		if frelay != nil {
			break
		}
	}

	// Sollte kein Relay gefunden wurden sein, wird der Vorgang abgebrochen
	if frelay == nil {
		obj._lock.Unlock()
		return nil, false, nil
	}

	// Der Threadlock wird feigegeben
	obj._lock.Unlock()

	// Der Vorgang wurde ohne fehler abgeschlossen
	return frelay, true, nil
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

			// Es wird geprüft ob die Routen bereits initalisiert wurden
			if obj._connection[i].RoutestInited() {
				log.Println("ConnectionManager: relay always inited =", relay._hexed_id)
				obj._lock.Unlock()
				return nil
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

// Erstellt einen neuen Verbindungs Manager
func newConnectionManager() ConnectionManager {
	return ConnectionManager{_connection: make([]*connection_io_pair, 0), _lock: new(sync.Mutex)}
}
