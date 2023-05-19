package kernel

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	"github.com/fluffelpuff/RoueX/kernel/extra"
)

// Stellt einen Relay Eintrag dar
type RelayConnectionEntry struct {
	_lock        *sync.Mutex
	_route_links []string
	PingTime     []uint64
	RelayLink    Relay
	Connections  []RelayConnection
}

// Gibt an, ob es sich um die gleiche Verbindung handelt
func (obj *RelayConnectionEntry) Equal(p2 *Relay) bool {
	return bytes.Equal(obj.RelayLink.GetPublicKey().SerializeCompressed(), p2.GetPublicKey().SerializeCompressed())
}

// Gibt den Hashwert des Objekts zurück
func (obj *RelayConnectionEntry) Hash() uint32 {
	var hash uint32
	for _, c := range obj.RelayLink.GetPublicKey().SerializeCompressed() {
		hash = 31*hash + uint32(c)
	}
	return hash
}

// Gibt an ob der Relay mindestens eine Aktive Verbindung hat
func (obj *RelayConnectionEntry) HasActiveConnection() bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird gerüft ob es mindestens eine Aktive Verbindung gibt
	for i := range obj.Connections {
		if obj.Connections[i].IsFinally() {
			if obj.Connections[i].IsConnected() {
				return true
			}
		}
	}

	// Es ist keine Aktive Verbindung vorhanden
	return false
}

// Gibt an ob es sich um eine bekannte Verbindung handelt
func (obj *RelayConnectionEntry) ConnectionIsKnown(conn RelayConnection) bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob es eine bekannte Verbindung gibt
	for i := range obj.Connections {
		if obj.Connections[i].GetObjectId() == conn.GetObjectId() {
			if obj.Connections[i].IsConnected() {
				return true
			} else {
				return false
			}
		}
	}

	// Es handelt sich um eine Unbekannte Verbindung
	return false
}

// Fügt dem Eintrag eine neue Verbindung hinzu
func (obj *RelayConnectionEntry) AddConnection(conn RelayConnection) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob es eine bekannte Verbindung gibt
	for i := range obj.Connections {
		if obj.Connections[i].GetObjectId() == conn.GetObjectId() {
			if obj.Connections[i].IsConnected() {
				return fmt.Errorf("AddConnection: 1: connection alrady added")
			}
		}
	}

	// Die Verbindung wird zwischengespeichert
	obj.Connections = append(obj.Connections, conn)

	// Log
	log.Println("RelayConnectionEntry: relay connection added. relay =", obj.RelayLink.GetPublicKeyHexString(), "connection =", conn.GetObjectId())
	return nil
}

// Entfernt einen Verbindungseintrag
func (obj *RelayConnectionEntry) RemoveConnection(conn RelayConnection) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob es eine bekannte Verbindung gib
	found_i := -1
	for i := range obj.Connections {
		if obj.Connections[i].GetObjectId() == conn.GetObjectId() {
			found_i = i
			break
		}
	}

	// Die Verbindung wird entfernt
	if found_i > -1 {
		obj.Connections = append(obj.Connections[:found_i], obj.Connections[found_i+1:]...)
	}

	// Log
	log.Println("RelayConnectionEntry: relay connection removed. relay =", obj.RelayLink.GetPublicKeyHexString(), "connection =", conn.GetObjectId())
	return nil
}

// Gibt die Gesamtzahl aller Verfügabren Verbindungen zurück
func (obj *RelayConnectionEntry) GetTotalConenctions() uint64 {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Das Ergebniss wird zurückgegeben
	return uint64(len(obj.Connections))
}

// Gibt die Metadaten des Relay Eintrags zurück
func (obj *RelayConnectionEntry) GetAllMetaInformationsOfRelayConnections() []RelayConnectionMetaData {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Die Metadaten werden erstellt

	return nil
}

// Gibt an wieviele Daten zum jetzigen Zeitpunk gesendet wurden
func (obj *RelayConnectionEntry) GetTotalWritedBytes() uint64 {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es werden alle Daten zusammengerechnet
	total := uint64(0)
	for i := range obj.Connections {
		tx, _ := obj.Connections[i].GetTxRxBytes()
		total += tx
	}

	// Die Gesamtmenge wird zurückgegebn
	return total
}

// Gibt an weiviele Daten zum jetzigen Zeitpunkt Empfangen wurden
func (obj *RelayConnectionEntry) GetTotalReadedBytes() uint64 {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es werden alle Daten zusammengerechnet
	total := uint64(0)
	for i := range obj.Connections {
		_, rx := obj.Connections[i].GetTxRxBytes()
		total += rx
	}

	// Die Gesamtmenge wird zurückgegebn
	return total
}

// Gibt den Durchschnittlichen Ping an
func (obj *RelayConnectionEntry) GetAveragePingTimeMS() uint64 {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Die Durchschnittlichen Zeiten werden zusammengerechnet
	p_time, c := uint64(0), uint64(0)
	for i := range obj.Connections {
		p_time += obj.Connections[i].GetPingTime()
		c++
	}

	// Die Durchschnittliche Zeit wird ermittelt
	result := uint64(p_time / c)

	// Die Daten werden zurückgegebn
	return result
}

// Gibt an ob es sich um eine Vertrauenswürdige Verbindung handelt
func (obj *RelayConnectionEntry) IsTrustedConnection() bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Gibt an ob die Verbindung vertrauenswürdig ist
	return obj.RelayLink.IsTrusted()
}

// Gibt alle Ausgehenden Verbindungen an
func (obj *RelayConnectionEntry) GetOutboundConnections() []RelayConnection {
	return nil
}

// Gibt alle Eingehenden Verbindungen aus
func (obj *RelayConnectionEntry) GetInbouncConnections() []RelayConnection {
	return nil
}

// Wird ausgeführt wenn der Kernel Signalisiert dass die Verbindung getrennt werden soll
func (obj *RelayConnectionEntry) DestroyByKernel() {
	obj._lock.Lock()
	defer obj._lock.Unlock()
	for i := range obj.Connections {
		obj.Connections[i].CloseByKernel()
	}
	return
}

// Gibt an ob die Routing Liste für diesen Relay bereits zugewiesen wurde
func (obj *RelayConnectionEntry) HasActiveRouteList() bool {
	return false
}

// registriert eine Routing Liste für diesen Relay
func (obj *RelayConnectionEntry) RegisterRouteList(rlist *RelayRoutesList) bool {
	return false
}

// Nimmt Pakete entgegen welche gesendet werden sollen
func (obj *RelayConnectionEntry) BufferL2PackageAndWrite(pckg *EncryptedAddressLayerPackage) (*extra.PackageSendState, error) {
	fmt.Println("SEND_PING")
	sstate := extra.NewPackageSendState()
	return sstate, nil
}
