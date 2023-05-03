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
	_lock            *sync.Mutex
	_relay           *Relay
	_routes_inited   bool
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

// Diese Funktion gibt an ob für diese Relay Verbindung bereits die Routen Initalisiert wurden
func (obj *connection_io_pair) RoutestInited() bool {
	obj._lock.Lock()
	r := obj._routes_inited
	obj._lock.Unlock()
	return r
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
	if obj._routes_inited {
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
	obj._routes_inited = true

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Log
	log.Println("connection_io_pair: relay and route activated. relay =", obj._relay._hexed_id, "address =", hex.EncodeToString(obj._relay._public_key.SerializeCompressed()))

	// Der Vorgang wurde ohne fehler durchgeführt
	return nil
}

// Erstellt ein neues Connection IO Pair
func createNewConnectionIoPair(relay *Relay) *connection_io_pair {
	return &connection_io_pair{_relay: relay, _lock: new(sync.Mutex), _routes_inited: false}
}
