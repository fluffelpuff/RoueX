package kernel

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"

	addresspackages "github.com/fluffelpuff/RoueX/address_packages"
	"github.com/fluffelpuff/RoueX/kernel/extra"
	"github.com/fluffelpuff/RoueX/rerror"
)

// Stellt einen Relay Eintrag dar
type RelayConnectionEntry struct {
	_lock            *sync.Mutex
	_route_list      *RelayRoutesList
	_signal_shutdown bool
	_closed          bool
	PingTime         []uint64
	RelayLink        Relay
	Connections      []RelayConnection
}

// Gibt an, ob es sich um die gleiche Verbindung handelt
func (obj *RelayConnectionEntry) Equal(p2 *Relay) bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob die zwei Öffentlichen Schlüssel übereinstimmen
	return bytes.Equal(obj.RelayLink.GetPublicKey().SerializeCompressed(), p2.GetPublicKey().SerializeCompressed())
}

// Gibt den Hashwert des Objekts zurück
func (obj *RelayConnectionEntry) Hash() uint32 {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Der Hash wird erstellt
	var hash uint32
	for _, c := range obj.RelayLink.GetPublicKey().SerializeCompressed() {
		hash = 31*hash + uint32(c)
	}

	// Der Hash wird zurückgegeben
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
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es werden alle Ausgehenden Verbindungen Extrahiert
	result := make([]RelayConnection, 0)
	for i := range obj.Connections {
		if obj.Connections[i].GetIOType() == OUTBOUND {
			if obj.Connections[i].IsConnected() && obj.Connections[i].IsFinally() {
				result = append(result, obj.Connections[i])
			}
		}
	}

	// Die Daten werden zurückgegeben
	return result
}

// Gibt alle Eingehenden Verbindungen aus
func (obj *RelayConnectionEntry) GetInbouncConnections() []RelayConnection {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es werden alle Ausgehenden Verbindungen Extrahiert
	result := make([]RelayConnection, 0)
	for i := range obj.Connections {
		if obj.Connections[i].GetIOType() == INBOUND {
			if obj.Connections[i].IsConnected() && obj.Connections[i].IsFinally() {
				result = append(result, obj.Connections[i])
			}
		}
	}

	// Die Daten werden zurückgegeben
	return result
}

// Gibt die Metadaten des Relay Eintrags zurück
func (obj *RelayConnectionEntry) GetAllMetaInformationsOfRelayConnections() []RelayConnectionMetaData {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Die Metadaten werden erstellt

	return nil
}

// Wird ausgeführt wenn der Kernel Signalisiert dass die Verbindung getrennt werden soll
func (obj *RelayConnectionEntry) CloseByKernel() {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Die Verbindungen werden geschlossen
	for i := range obj.Connections {
		go obj.Connections[i].CloseByKernel()
	}

	// Es wird Signalisiert dass das Objekt beendet werden soll
	obj._signal_shutdown = true

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Es wird gewartet bis alle Verbindungen geschlossen wurden
	for obj.HasActiveConnection() {
		time.Sleep(1 * time.Millisecond)
	}

	// Der Threadlock wird final angewendet
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird festgelegt dass das Objekt erfolgreich geschlossen wurde
	obj._closed = true
}

// Gibt an ob die Routing Liste für diesen Relay bereits zugewiesen wurde
func (obj *RelayConnectionEntry) HasActiveRouteList() bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Das Ergebniss wirdzurückgegeben
	return obj._route_list != nil
}

// registriert eine Routing Liste für diesen Relay
func (obj *RelayConnectionEntry) RegisterRouteList(rlist *RelayRoutesList) bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob bereits eine Routing Liste gesetzt wurde
	if obj._route_list != nil {
		return false
	}

	// Die Routingliste wird zwischengespeichert
	obj._route_list = rlist

	// Der Vorgang wurde erfolgreich druchgeführt
	return true
}

// Nimmt Pakete entgegen welche gesendet werden sollen
func (obj *RelayConnectionEntry) BufferL2PackageAndWrite(pckg *addresspackages.FinalAddressLayerPackage) (*extra.PackageSendState, error) {
	// Es wird geprüft ob eine Aktive Verbindung verfügbar ist
	if obj.HasActiveConnection() {
		return nil, fmt.Errorf("no active connection for this route")
	}

	// Das Paket wird in Bytes umgewandelt
	byted_pckge, err := pckg.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("BufferL2PackageAndWrite: " + err.Error())
	}

	// Das Rückgabe Objekt wird erstellt
	sstate := extra.NewPackageSendState()

	// Es wird nach einer passenden Verbindung gesucht
	var found_conn RelayConnection
	for i := 0; i < 2; i++ {
		// Im ersten verusch werden alle eingehenden Verbindungen abgerufen
		// im zweiten durchgang werden eingehende Verbindung abgerufen
		var clist []RelayConnection
		if i == 0 {
			clist = obj.GetOutboundConnections()
		} else {
			clist = obj.GetInbouncConnections()
		}

		// Sollte die CList leer sein wird ein Pani ausgelöst
		if clist == nil {
			panic("BufferL2PackageAndWrite: unkown error")
		}

		// Es wird geprüft ob eine Verbindung verfügbar ist
		if len(clist) < 1 {
			continue
		}

		// Es wird eine Verfügabre Verbindung herausgesucht
		for x := range clist {
			// Es wird geprüft ob die ausgewählte Verbindung initalisiert und fertigestellt wurde
			if !clist[x].IsFinally() || !clist[x].IsConnected() {
				continue
			}

			// Es wird geprüft ob die Verbindung zum schreiben verwendet werden kann
			if !clist[x].CannUseToWrite() {
				continue
			}

			// Speichert die gefundene Verbindung zwischen
			found_conn = clist[x]
			break
		}

		// Es wird geprüft ob eine Verbindung verfügabr ist, wenn ja wird die Schleife abgebrochen
		if found_conn != nil {
			break
		}
	}

	// Sollte keine Verbindung vorhanden sein, wird der Vorgang abgebrochen
	if found_conn == nil {
		return nil, fmt.Errorf("")
	}

	// Die Daten werden an die Verbindung übergeben
	if ste, err := found_conn.EnterSendableData(byted_pckge, sstate); err != nil || !ste {
		// Der Status wird auf DROPED gesetzt
		sstate.SetFinallyState(extra.DROPED)

		// Es wird geprüft ob ein Fehler aufgetreten ist
		if err != nil {
			if found_conn.IsConnected() {
				return sstate, fmt.Errorf("BufferL2PackageAndWrite: " + err.Error())
			} else {
				return sstate, &rerror.IOStateError{}
			}
		}

		// Es wird geprüft ob die Verbindung mit der ausgewählten Verbindung noch besteht
		if !found_conn.IsConnected() {
			return sstate, &rerror.IOStateError{}
		}

		// Wenn die Verbindung noch besteht, wurde das Paket aufgrund eines vollen buffers verworfen
		return sstate, nil
	}

	// Das Paket wurde erfolreich an den Verbindungspuffer übergeben
	return sstate, nil
}
