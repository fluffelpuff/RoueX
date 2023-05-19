package kernel

import (
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fluffelpuff/RoueX/kernel/extra"
)

// Stellt den Verbindungsmanager dar
type RelayConnectionRoutingTable struct {
	_connection_relay_map map[string]*Relay
	_route_ro_relay       map[string]*RelayConnectionEntry
	_relays_map           map[*Relay]*RelayConnectionEntry
	_lock                 *sync.Mutex
	_is_closed            bool
}

// Gibt an ob es eine Aktive Verbindung gibt
func (obj *RelayConnectionRoutingTable) HasActiveConnections() bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob es eine Aktive Verbindung gibt
	if len(obj._relays_map) == 0 {
		return false
	}

	// Es wird ermittelt ob es mindestens eine Aktive Verbindung gibt
	for i := range obj._relays_map {
		for x := range obj._relays_map[i].Connections {
			if obj._relays_map[i].Connections[x].IsConnected() {
				return true
			}
		}
	}

	// Es wurde keine Verbindung gefunden
	return false
}

// Fügt eine neue AKtive Verbindung zum Manager hinzu
func (obj *RelayConnectionRoutingTable) RegisterNewRelayConnection(relay *Relay, conn RelayConnection) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob die Routing Tabelle geschlossen wurde
	if obj._is_closed {
		return fmt.Errorf("RegisterNewRelayConnection: 1: routing table are closed")
	}

	// Es wird ermittelt ob es bereits einen Eintrag für diesen Relay gibt
	relay_entry, found_relay_entry := obj._relays_map[relay]
	if !found_relay_entry {
		// Der Relay Eintrag wird erzeugt
		relay_entry = &RelayConnectionEntry{
			RelayLink:    *relay,
			Connections:  []RelayConnection{},
			PingTime:     []uint64{},
			_lock:        new(sync.Mutex),
			_route_links: []string{},
		}

		// Der Eintrag wird abgespeichert
		obj._relays_map[relay] = relay_entry

		// Der Relay Connection Entry wird der Adresse Route zugewiesen
		obj._route_ro_relay[hex.EncodeToString(relay._public_key.SerializeCompressed())] = relay_entry

		// Log
		log.Println("RelayConnectionRoutingTable: new relay and route added. relay =", relay.GetPublicKeyHexString())
	} else {
		// Es wird geprüft ob die Verbindung dem Relay bereits zugewiesen wurde
		if relay_entry.ConnectionIsKnown(conn) {
			return fmt.Errorf("RemoveConnectionFromRelay: 1: the connection always added")
		}
	}

	// Es wird versucht die Verbindung in dem Relay hinzuzufügen
	if err := relay_entry.AddConnection(conn); err != nil {
		return fmt.Errorf("RegisterNewRelayConnection: 2: " + err.Error())
	}

	// Die VerbindungsID wird dem Relay Eintrag zugewiesen
	obj._connection_relay_map[conn.GetObjectId()] = relay

	// Der Vorgang wurde ohne Fehler erfolgreich druchgeführt
	return nil
}

// Entfernt eine Verbindung
func (obj *RelayConnectionRoutingTable) RemoveConnectionFromRelay(conn RelayConnection) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob die Routing Tabelle geschlossen wurde
	if obj._is_closed {
		return fmt.Errorf("RemoveConnectionFromRelay: 1: routing table are closed")
	}

	// Es wird geprüft ob es für diese Verbindung einen Eintrag gibt
	relay_link, has_found := obj._connection_relay_map[conn.GetObjectId()]
	if !has_found {
		return fmt.Errorf("RemoveConnectionFromRelay: 1: the connection cant found")
	}

	// Der Relay eintrag wird ermittelt
	relay_entry, has_found := obj._relays_map[relay_link]
	if !has_found {
		return fmt.Errorf("RemoveConnectionFromRelay: 2: the connection cant found")
	}

	// Die Verbindung wird aus dem Relay Eintrag entfernt
	if err := relay_entry.RemoveConnection(conn); err != nil {
		return fmt.Errorf("RemoveConnectionFromRelay: 3: " + err.Error())
	}

	// Die Verlinkung mit dem Relay wird entfernt
	delete(obj._connection_relay_map, conn.GetObjectId())

	// Es wird geprüft ob dem Relay noch weitere Verbindungen zugeodnet sind
	if len(relay_entry.Connections) < 1 {
		// Log
		log.Println("RelayConnectionRoutingTable: relay complete removed. relay =", relay_link.GetPublicKeyHexString())

		// Der Eintrag wird entfernt
		delete(obj._relays_map, relay_link)
	}

	// Der Vorgang wurde ohne fehler erfolgreich durchgeführt
	return nil
}

// Gibt an ob der Relay Verbunden ist
func (obj *RelayConnectionRoutingTable) RelayIsConnected(relay *Relay) bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob es einen Eintrag für diesen Relay gibt
	relay_entry, found := obj._relays_map[relay]
	if !found {
		return false
	}

	// Die Antwort wird zurückgegeben
	return relay_entry.HasActiveConnection()
}

// Gibt an weiviele Verbindungen ein Relay hat
func (obj *RelayConnectionRoutingTable) GetTotalRelayConnections(relay *Relay) uint64 {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob es einen Eintrag für diesen Relay gibt
	relay_entry, found := obj._relays_map[relay]
	if !found {
		return 0
	}

	// Die Antwort wird zurückgegeben
	return relay_entry.GetTotalConenctions()
}

// Ruft alle MetaDaten über die Verbindungen eines Relays ab
func (obj *RelayConnectionRoutingTable) GetAllMetaInformationsOfRelayConnections(relay *Relay) (*RelayMetaData, error) {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob die Routing Tabelle geschlossen wurde
	if obj._is_closed {
		return nil, fmt.Errorf("GetAllMetaInformationsOfRelayConnections: 1: routing table are closed")
	}

	// Es wird geprüft ob es einen Eintrag für diesen Relay gibt
	entry, found := obj._relays_map[relay]
	if !found {
		return nil, fmt.Errorf("GetAllMetaInformationsOfRelayConnections: 1: unkown relay")
	}

	// Die Antwort wird gebaut
	result := &RelayMetaData{
		Connections: entry.GetAllMetaInformationsOfRelayConnections(),
		PublicKey:   relay.GetPublicKeyHexString(),
		IsConnected: entry.HasActiveConnection(),
		TotalWrited: uint64(entry.GetTotalWritedBytes()),
		TotalReaded: uint64(entry.GetTotalReadedBytes()),
		PingMS:      entry.GetAveragePingTimeMS(),
		IsTrusted:   entry.IsTrustedConnection(),
	}

	// Die Antwort wird zurückgegeben
	return result, nil
}

// Ruft ein Relay anhand einer Verbindung ab
func (obj *RelayConnectionRoutingTable) GetRelayByConnection(conn RelayConnection) (*Relay, bool, error) {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob die Routing Tabelle geschlossen wurde
	if obj._is_closed {
		return nil, false, fmt.Errorf("GetRelayByConnection: 1: routing table are closed")
	}

	// Es wird geprüft ob es für diese Verbindung einen Eintrag gibt
	relay_link, has_found := obj._connection_relay_map[conn.GetObjectId()]
	if !has_found {
		return nil, false, nil
	}

	// Die Antwort wird zurückgegeben
	return relay_link, true, nil
}

// Wird verwendet um alle Aktiven Routen für ein Relay zu Initalisieren
func (obj *RelayConnectionRoutingTable) InitRoutesForRelay(rlay *Relay, routes *RelayRoutesList) (bool, bool) {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob die Routing Tabelle geschlossen wurde
	if obj._is_closed {
		return false, false
	}

	// Es wird geprüft ob ein passender Relay vorhanden ist
	relay, has_found := obj._relays_map[rlay]
	if !has_found {
		return false, false
	}

	// Es wird geprüft ob die Routing Liste schon zu gewiesen wurde
	if relay.HasActiveRouteList() {
		return true, false
	}

	// Die Routingliste wird dem Relay zugewiesen
	if ok := relay.RegisterRouteList(routes); !ok {
		if relay.HasActiveConnection() {
			return true, false
		} else {
			return false, false
		}
	}

	// Das Relay wird herausgesucht
	return true, true
}

// Wird vom Kernel verwendet alle Verbindungen zu schließen
func (obj *RelayConnectionRoutingTable) ShutdownByKernel() {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird allen Relays Signalisiert dass sie all ihre Verbindungen schließen sollen
	for i := range obj._relays_map {
		obj._relays_map[i].DestroyByKernel()
	}

	// Es wird Signalisiert dass das Objekt gehschlossen werden soll
	obj._is_closed = true

	// Der Threadlock wird freigeben
	obj._lock.Unlock()

	// Die Verbindungen werden geschlossen

	// Es wird gewartet bis alle Verbindungen geschlossen wurden
	for obj.HasActiveConnections() {
		time.Sleep(1 * time.Millisecond)
	}
}

// Nimmt Pakete entgegen und Routet diese zu dem Entsprechenden Host
func (obj *RelayConnectionRoutingTable) EnterPackageToRoutingManger(pckg *EncryptedAddressLayerPackage) (*extra.PackageSendState, bool, error) {
	// Der Threadlock wird verwnendet
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob die Routing Tabelle geschlossen wurde
	if obj._is_closed {
		return nil, false, fmt.Errorf("EnterPackageToRoutingManger: 1: routing table are closed")
	}

	// Das Passende Relay für diese Verbindung wird herausgefiltert
	route_ep, found_route := obj._route_ro_relay[hex.EncodeToString(pckg.Reciver.SerializeCompressed())]
	if !found_route {
		return nil, false, nil
	}

	// Es wird geprüft ob eine Route samt Verbindung gefunden wurde
	if !found_route {
		return nil, false, nil
	}

	// Es wird geprüft ob mit der Gegenseite ein Verbindung besteht
	if !route_ep.HasActiveConnection() {
		return nil, false, nil
	}

	// Das Paket wird an die Verbindung übergeben
	sstate, err := route_ep.BufferL2PackageAndWrite(pckg)
	if err != nil {
		return nil, true, fmt.Errorf("EnterPackageToRoutingManger: 2: " + err.Error())
	}

	// Das Paket wurde erfolgreich an die Verbindung gesendet
	return sstate, true, nil
}

// Erstellt einen neuen Verbindungs Manager
func newRelayConnectionRoutingTable() RelayConnectionRoutingTable {
	return RelayConnectionRoutingTable{
		_connection_relay_map: make(map[string]*Relay),
		_route_ro_relay:       make(map[string]*RelayConnectionEntry),
		_relays_map:           make(map[*Relay]*RelayConnectionEntry),
		_lock:                 new(sync.Mutex),
		_is_closed:            false,
	}
}
