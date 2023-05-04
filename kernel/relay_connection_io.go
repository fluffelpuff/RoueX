package kernel

import (
	"encoding/hex"
	"fmt"
	"log"
	"sync"
)

// Stellt eine Verbindungspaar dar
type connection_io_pair struct {
	_route_endpoints []RouteEntry
	_conn            []RelayConnection
	_lock            sync.Mutex
	_relay           *Relay
	_active          bool
}

// Diese Funktion gibt an ob es eine Aktive verbindung für diesen Relay gibt
func (obj *connection_io_pair) HasActiveConnections() bool {
	obj._lock.Lock()
	for i := range obj._conn {
		if obj._conn[i].IsConnected() {
			obj._lock.Unlock()
			return true
		}
	}
	obj._lock.Unlock()
	return false
}

// Fügt dem Paar eine Verbindung hinzu
func (obj *connection_io_pair) add_connection(conn RelayConnection) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob die Verbindung diesem Relay bereits zugeodnet wurde
	for i := range obj._conn {
		if obj._conn[i].GetObjectId() == conn.GetObjectId() {
			obj._lock.Unlock()
			return fmt.Errorf("connection_io_pair: add_connection: 1: connection always added")
		}
	}

	// Die Verbindung wird hinzugefügt
	obj._conn = append(obj._conn, conn)

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Log
	log.Println("connection_io_pair: new connection to relay added. relay =", obj._relay._hexed_id, "connection =", conn.GetObjectId(), "protocol =", conn.GetProtocol())

	// Der Vorgang wurde ohne fehler durchgeführt
	return nil
}

// Entfernt eine Spezifische Verbindung
func (obj *connection_io_pair) remove_connection(conn RelayConnection) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob die Verbindung diesem Relay bereits zugeodnet wurde
	to_remove_hight := -1
	for i := range obj._conn {
		if obj._conn[i].GetObjectId() == conn.GetObjectId() {
			to_remove_hight = i
			break
		}
	}

	// Sollte die Verbindung nicht gefunden wurden sein, wird der Vorgang abgebrochen
	if to_remove_hight == -1 {
		obj._lock.Unlock()
		return nil
	}

	// Die Verbindung wird entfernt
	obj._conn = append(obj._conn[:to_remove_hight], obj._conn[to_remove_hight+1:]...)

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Log
	log.Println("connection_io_pair: connection from relay removed. relay =", obj._relay._hexed_id, "connection =", conn.GetObjectId(), "protocol =", conn.GetProtocol())

	// Es wird geprüft ob noch eine weitere Verbindung vorhanden ist, wenn nein wird der Relay deaktiviert
	if !obj.HasActiveConnections() {
		obj._event_no_active_connection()
	}

	// Der Vorgang wurde ohne fehler durchgeführt
	return nil
}

// Fügt Routen hinzu
func (obj *connection_io_pair) _add_route_entrys(routes []*RouteEntry) error {
	// Der Threadlock wird verwendet
	obj._lock.Lock()

	// Die Routen welche hinzugefügt werden soll
	for i := range routes {
		// Es wird geprüft ob die Route bereits vorhanden ist
		has_found_route_endpoint := false
		for x := range obj._route_endpoints {
			if obj._route_endpoints[x]._relay_hex_id == routes[i]._relay_hex_id {
				has_found_route_endpoint = true
				break
			}
		}

		// Sollte der Eintrag gefunden wurden sein, wird
		if has_found_route_endpoint {
			continue
		}

		// Log
		log.Println("connection_io_pair: add route to relay =", obj._relay._hexed_id, "address =", obj._relay._hexed_id)

		// Der Eintrag wird hinzugefügt
		obj._route_endpoints = append(obj._route_endpoints, *routes[i])
	}

	// Gibt den Threadlock frei
	obj._lock.Unlock()

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

// Signalisiert dass der Relay verwendet werden kann
func (obj *connection_io_pair) _signal_activated() error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob die Routen bereits initalisiert wurden
	if obj._active {
		obj._lock.Unlock()
		return nil
	}

	// Es wird geprüft ob eine Aktive Verbindung vorhanden ist
	has_found_active_connection := false
	for i := range obj._conn {
		if obj._conn[i].IsConnected() {
			has_found_active_connection = true
			break
		}
	}

	// Sollte keine Aktive Verbindung vorhanden sein wird der Vorgang abgebrochen
	if !has_found_active_connection {
		obj._lock.Unlock()
		return fmt.Errorf("connection_io_pair: relay has no active connection. relay =" + obj._relay._hexed_id)
	}

	// Das Relay wird Aktiviert
	obj._active = true

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Log
	log.Println("connection_io_pair: relay and route activated. relay =", obj._relay._hexed_id, "address =", hex.EncodeToString(obj._relay._public_key.SerializeCompressed()))

	// Der Vorgang wurde ohne fehler durchgeführt
	return nil
}

// Wird ausgeführt wenn keine Verbindung mehr vorhanden ist
func (obj *connection_io_pair) _event_no_active_connection() {
	// Der Threadlock wird verwendet
	obj._lock.Lock()

	// Es wird nocheinmal geprüft ob eine Verbindung vorhanden ist
	if len(obj._conn) != 0 {
		obj._lock.Unlock()
		return
	}

	// Es wird geprüft ob das Relay beits Deaktiviert wurde
	if !obj._active {
		obj._lock.Unlock()
		return
	}

	// Das Objekt wird Deaktiviert
	obj._active = false

	// Der Threadlock wird freigebene
	obj._lock.Unlock()

	// Gibt den Log an
	log.Println("connection_io_pair: relay has no connections, disabled. relay =", obj._relay._hexed_id)
}

// Gibt an ob die Verbindung teil das Paares ist
func (obj *connection_io_pair) _is_used_conenction(conn RelayConnection) bool {
	// Der Threadlock wird verwendet
	obj._lock.Lock()

	// Es wird geprüft ob die Verbindung von diesem Relay genutzt wird
	is_used := false
	for i := range obj._conn {
		if obj._conn[i].GetObjectId() == conn.GetObjectId() {
			is_used = true
			break
		}
	}

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Der Status wird zurückgegeb
	return is_used
}

// Wird verwendet um alle Verbindungen durch den Kernel zu schließen
func (obj *connection_io_pair) kernel_shutdown() {
	obj._lock.Lock()
	cp_list := obj._conn
	obj._lock.Unlock()
	for i := range cp_list {
		cp_list[i].CloseByKernel()
	}
}

// Erstellt ein neues Connection IO Pair
func createNewConnectionIoPair(relay *Relay) *connection_io_pair {
	return &connection_io_pair{_relay: relay, _lock: sync.Mutex{}, _active: false}
}
