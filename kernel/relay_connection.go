package kernel

import (
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
)

// Stellt eine Verbindungspaar dar
type connection_io_pair struct {
	_route_endpoints []RoutingManagerEntry
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
func (obj *connection_io_pair) _add_route_entrys(routes []*RoutingManagerEntry) error {
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

// Ermittelt die Schnellste Verbindung
func (obj *connection_io_pair) GetBestConnection() RelayConnection {
	// Der Threadlock wird verwendet
	obj._lock.Lock()

	// Die Verbindungen werden Sortiert
	pre_sorted := []RelayConnection(obj._conn)
	sort.Sort(ByRelayConnection(pre_sorted))

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Es wird geprüft ob Mindestens eine Verbindung verfügbar ist
	if len(pre_sorted) < 1 {
		return nil
	}

	// Es wird geprüft ob die Schnellste Verbindung noch vorhanden ist
	for len(pre_sorted) > 0 {
		if pre_sorted[0].IsConnected() {
			return pre_sorted[0]
		} else {
			pre_sorted = pre_sorted[1:]
		}
	}

	// Es ist keine Verbindung verfügbar
	return nil
}

// Gibt die Gesendete und Empfange Daten Menge an
func (obj *connection_io_pair) GetRxTxBytes() (uint64, uint64) {
	// Es werden alle Verbindungen abgerufen
	obj._lock.Lock()
	conns := []RelayConnection(obj._conn)
	obj._lock.Unlock()

	// Solte keine Verbindung vorhanden sein
	if len(conns) < 1 {
		return 0, 0
	}

	// Die Gesamte Datenmenge wird ermittelt
	tx, rx := uint64(0), uint64(0)
	for i := range conns {
		if !conns[i].IsConnected() {
			continue
		}
		_tx, _rx := conns[i].GetTxRxBytes()
		tx += _tx
		rx += _rx
	}

	// Die Daten werden zurückgegeben
	return tx, rx
}

// Gibt an wieviele Verbindungen aufgebaut sind
func (obj *connection_io_pair) GetTotalConnection() uint64 {
	obj._lock.Lock()
	r := uint64(len(obj._conn))
	obj._lock.Unlock()
	return r
}

// Ruft alle Verfügabren Verbindungen ab
func (obj *connection_io_pair) GetConnections() []RelayConnection {
	obj._lock.Lock()
	r := []RelayConnection(obj._conn)
	obj._lock.Unlock()
	return r
}

// Gibt die Aktuellen MetaDaten der Verbindung uas
func (obj *connection_io_pair) GetMetaDataInformations() (*RelayMetaData, error) {
	// Die Schnelleste Verbindung wird ermittelt
	best_conn := obj.GetBestConnection()
	if best_conn == nil {
		return &RelayMetaData{
			TotalConnections: 0,
			IsConnected:      false,
			IsTrusted:        obj._relay._trusted,
			TotalWrited:      0,
			TotalReaded:      0,
			PingMS:           0,
			BandwithKBs:      0,
		}, nil
	}

	// Gibt die Gesendete und Empfange Datenmenge an
	tx_bytes, rx_bytes := obj.GetRxTxBytes()

	// Es werden alle Verbindungen abgerufen
	connection := obj.GetConnections()

	// Die MetaDaten werden abgerufen
	revals := make([]RelayConnectionMetaData, 0)
	for i := range connection {
		rx, tx := connection[i].GetTxRxBytes()
		pkey, err := connection[i].GetSessionPKey()
		if err != nil {
			return nil, err
		}
		encoded := hex.EncodeToString(pkey.SerializeCompressed())
		robj := RelayConnectionMetaData{
			Id:              connection[i].GetObjectId(),
			Protocol:        connection[i].GetProtocol(),
			InboundOutbound: uint8(connection[i].GetIOType()),
			IsConnected:     connection[i].IsConnected(),
			SessionPKey:     encoded,
			RxBytes:         rx,
			TxBytes:         tx,
		}
		if !connection[i].IsConnected() {
			continue
		}
		revals = append(revals, robj)
	}

	//Erzeugt den Rückgabewert
	result := &RelayMetaData{
		PingMS:           best_conn.GetPingTime(),
		IsTrusted:        obj._relay._trusted,
		TotalWrited:      tx_bytes,
		TotalReaded:      rx_bytes,
		IsConnected:      best_conn.IsConnected(),
		TotalConnections: uint64(len(revals)),
		Connections:      revals,
		BandwithKBs:      0,
	}

	// Gibt die Daten zurück
	return result, nil
}

// Gibt an, ob dieser Relay eine Route für eine bestimmte Adresse hat
func (obj *connection_io_pair) HasRouteForAddress(pkey *btcec.PublicKey) bool {
	// Es wird geprüft ob es eine Aktive verbindung gibt
	if !obj.HasActiveConnections() {
		return false
	}

	return false
}

// Nimmt Pakete entgegen und leitet sie an die beste Verbindung weiter
func (obj *connection_io_pair) EnterAndForwardPlainAddressLayerPackage(pckg *PlainAddressLayerPackage) (bool, error) {
	// Es wird geprüft ob es eine Aktive verbindung gibt
	if !obj.HasActiveConnections() {
		return false, nil
	}
	return false, nil
}

// Erstellt ein neues Connection IO Pair
func createNewConnectionIoPair(relay *Relay) *connection_io_pair {
	return &connection_io_pair{_relay: relay, _lock: sync.Mutex{}, _active: false}
}
