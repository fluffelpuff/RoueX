package kernelprotocols

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fluffelpuff/RoueX/utils"
	"github.com/fxamacker/cbor"
)

// Stellt ein Ping Paket dar
type PingPackage struct {
	Id string
}

// Gibt den Status eines Ping Vorganges an
type ping_state uint8

// Definiert alle verfügbaren Ping Vorgänge
const (
	ABORTED          = ping_state(0)
	RESPONDED        = ping_state(1)
	CLOSED_BY_KERNEL = ping_state(2)
	TIMEOUT          = ping_state(3)
	INTERNAL_ERROR   = ping_state(4)
)

// Stellt einen Ping Vorgangseintrag dar
type rouex_entry struct {
	_id            string
	_max_wait_time uint32
	_aborted       bool
	_start_time    time.Time
	_finally_time  *time.Time
	_kernel        *kernel.Kernel
	_lock          *sync.Mutex
}

// Gibt an ob die Maxiamle Zeit erreicht wurde
func (obj *rouex_entry) maxwnend() bool {
	// Die Zeitwerte werden abgerufen
	obj._lock.Lock()
	start_time := obj._start_time
	fin_time := obj._finally_time
	max_wait := obj._max_wait_time
	obj._lock.Unlock()

	// Die Ablaufzeit wird berechnet
	future_timeout := uint64(uint64(start_time.Nanosecond()/1000000) + uint64(max_wait))

	// Es wird ein Aktueller Zeitstempel erstellt
	current_timestamp := uint64(time.Now().Nanosecond() / 1000000)

	// Es wird ermittelt ob der Aktuelle Zeitstempel die Zukünftige Zeit überschreitet
	if future_timeout >= current_timestamp {
		return true
	}

	// Sollte die Finally Zeit vorhanden sein, wird der Vorgang abgebrochen
	if obj._finally_time != nil {
		// Es wird geprüft ob die Daten korrekt sind
		ms_finn_time := uint64(fin_time.Nanosecond() / 1000000)

		// Es wird geprüft ob die Aktuelle Zeit abgelaufen ist
		if ms_finn_time >= future_timeout {
			return true
		}
	}

	// Der Wert wird zurückgegeben
	return false
}

// Gibt an ob das Objekt fertigestellt wurde
func (obj *rouex_entry) isfinn() bool {
	// Die benötigten Variablen werden abgerufen
	obj._lock.Lock()
	sta := obj._finally_time
	obj._lock.Unlock()

	// Es wird geprüft ob die Zeit abgelaufen ist
	if obj.maxwnend() {
		return false
	}

	// Es wird geprüft ob der Vorgang abgebrochen wurde
	if obj.isaborted() {
		return false
	}

	// Sollte kein Zeitstempel vorhanden sein, wird ein False zurückgegeben
	if sta == nil {
		return false
	}

	// Der Vorgang wurde erfolgreich fertigestellt
	return true
}

// Gibt an ob das Objekt Abgebrochen wurde
func (obj *rouex_entry) isaborted() bool {
	// Der Aborted Wert wird abgerufen
	obj._lock.Lock()
	rst := obj._aborted
	obj._lock.Unlock()

	// Der Wert wird zurückgegeben
	return rst
}

// Diese Funktion wird solange ausgeführt, bis die Schleife beendet wurde
func (obj *rouex_entry) waitfnc() (ping_state, error) {
	// Diese Funktion gibt an ob der Aktuelle Vorgang nocht ausgeführt wird
	isvfnc := func(obj *rouex_entry) bool {
		// Es wird ermittelt ob der Kernel ausgeführt werden soll
		if !obj._kernel.IsRunning() {
			return false
		}

		// Es wird geprüft ob die Zeit abgelaufen ist
		if obj.maxwnend() {
			return false
		}

		// Es wird geprüft ob die benötigte
		if obj.isaborted() {
			return false
		}

		// Es wird geprüft ob dieser Vorgang bereits beantwortet wurde
		if obj.isfinn() {
			return false
		}

		// Gibt die Antwort zurück
		return true
	}

	// Diese Schleife wird solange ausgeführt bis der Ping Vorgang abgerlaufen ist oder beantwortet wurde
	for isvfnc(obj) {
		time.Sleep(1 * time.Millisecond)
	}

	// Der Aktuelle Status wird ermittelt
	if !obj._kernel.IsRunning() {
		return CLOSED_BY_KERNEL, nil
	}

	// Es wird geprüft ob die Zeit abgelaufen ist
	if obj.maxwnend() {
		return TIMEOUT, nil
	}

	// Es wird geprüft ob der Vorgang fertigesteltl wurde
	if obj.isaborted() {
		return ABORTED, nil
	}

	// Der Vorgang wurde abgebrochen
	return RESPONDED, nil
}

// Gibt die Benötigte Zeit an
func (obj *rouex_entry) gtimems() uint32 {
	// Die Daten werden mit dem Threadlock abgerufen wird
	obj._lock.Lock()
	stime := obj._start_time
	etime := obj._finally_time
	obj._lock.Unlock()

	// Sollte noch kein Endzeit vorhanden sein wird der Vorgang abgebrochen
	if etime == nil {
		return obj._max_wait_time
	}

	// Die benötigte Zeit wird ausgerechnet
	total_time := stime.Sub(*etime)

	// Sollte die benötigte Zeit größer als
	if total_time >= math.MaxUint32 {
		return math.MaxUint32
	}

	// Sollte die Maximale Zeit abgelaufen sein, wird diese Zurückgegeben
	if obj._max_wait_time >= uint32(total_time) {
		return obj._max_wait_time
	}

	// Die benötigte Zeit wird zurückgegeben
	return uint32(total_time.Milliseconds())
}

// Stellt das Ping Protokoll dar
type ROUEX_PING_PONG_PROTOCOL struct {
	_open_processes []rouex_entry
	_objid          string
	_kernel         *kernel.Kernel
	_lock           *sync.Mutex
}

// Fügt einen Ping Prozess hinzu
func (obj *ROUEX_PING_PONG_PROTOCOL) _add_ping_process(ping_proc rouex_entry) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Speichert den Vorgang ab
	obj._open_processes = append(obj._open_processes, ping_proc)

	// Der Threadlock wird freigegebeb
	obj._lock.Unlock()

	// Log
	log.Printf("ROUEX_PING_PONG_PROTOCOL: new ping process created. pid = %s\n", obj._objid)

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

// Entfernt einen Ping Prozess
func (obj *ROUEX_PING_PONG_PROTOCOL) _remove_ping_process(ping_proc rouex_entry) {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Speichert den Vorgang ab
	obj._open_processes = append(obj._open_processes, ping_proc)

	// Der Threadlock wird freigegebeb
	obj._lock.Unlock()

	// Log
	log.Printf("ROUEX_PING_PONG_PROTOCOL: ping process removed. pid = %s\n", obj._objid)
}

// Führt einen Ping Prozess durch
func (obj *ROUEX_PING_PONG_PROTOCOL) _start_ping_pong_process(pkey *btcec.PublicKey) (map[string]interface{}, error) {
	// Die Prozess ID wird erstellt
	proc_id := utils.RandProcId()

	// Das Paket wird gebaut
	builded_package := PingPackage{Id: proc_id}

	// Das Paket wird in Bytes umgewandelt
	encoded_ping_package, err := cbor.Marshal(builded_package, cbor.EncOptions{})
	if err != nil {
		panic(err)
	}

	// Es wird ein neuer Ping Prozess wird erzeugt
	rx_entry := rouex_entry{_id: proc_id, _lock: new(sync.Mutex), _kernel: obj._kernel, _start_time: time.Now()}

	// Der Vorgang wird abgespeichert
	if err := obj._add_ping_process(rx_entry); err != nil {
		return nil, fmt.Errorf("_start_ping_pong_process: " + err.Error())
	}

	// Das Ping Paket wird über das Netzwerk übermittelt
	has_route, err := obj._kernel.EnterBytesAndSendL2PackageToNetwork(0, encoded_ping_package, pkey)
	if err != nil {
		// Der Ping Prozess wird wieder entfernt
		obj._remove_ping_process(rx_entry)

		// Der Fehler wird zurückgegeben
		return nil, fmt.Errorf("_start_ping_pong_process: " + err.Error())
	}

	// Sollte keine Route vorhanden sein, wird der Vorgang abgebrochen
	if !has_route {
		// Der Pingvorgang wird entfernt
		obj._remove_ping_process(rx_entry)

		// Der Fehler wird zurückgegeben
		return nil, fmt.Errorf("no route found")
	}

	// Es wird auf die Antwort wird gewartet
	state, err := rx_entry.waitfnc()
	if err != nil {
		// Der Pingvorgang wird entfernt
		obj._remove_ping_process(rx_entry)

		// Der Fehler wird zurückgegeben
		return nil, fmt.Errorf("_start_ping_pong_process: " + err.Error())
	}

	// Der Pingvorgang wird entfernt
	obj._remove_ping_process(rx_entry)

	// Das Rückgabeobjekt wird erstellt
	reval := make(map[string]interface{})

	// Es wird geprüft das der Vorgang mit einem Response beantwortet wurde
	switch state {
	case ABORTED:
		log.Printf("ROUEX_PING_PONG_PROTOCOL: ping process aborted. pid = %s\n", proc_id)
		reval["state"] = uint8(1)
		return reval, nil
	case RESPONDED:
		log.Printf("ROUEX_PING_PONG_PROTOCOL: ping process responded. pid = %s\n", proc_id)
		reval["state"] = uint8(0)
		reval["ttime"] = rx_entry.gtimems()
		return reval, nil
	case CLOSED_BY_KERNEL:
		log.Printf("ROUEX_PING_PONG_PROTOCOL: ping process closed by kernel. pid = %s\n", proc_id)
		reval["state"] = uint8(2)
		return reval, nil
	default:
		return nil, fmt.Errorf("unkown state")
	}
}

// Nimmt eingetroffene Pakete aus dem Netzwerk Entgegen
func (obj *ROUEX_PING_PONG_PROTOCOL) EnterRecivedPackage(pckage *kernel.AddressLayerPackage, conn kernel.RelayConnection) error {
	return nil
}

// Nimmt Datensätze entgegen und übergibt diese an den Kernel um das Paket entgültig abzusenden
func (obj *ROUEX_PING_PONG_PROTOCOL) EnterWritableBytesToReciver(data []byte, reciver *btcec.PublicKey) error {
	return nil
}

// Nimmt eintreffende Steuer Befehele entgegen
func (obj *ROUEX_PING_PONG_PROTOCOL) EnterCommandData(command string, arguments [][]byte) (map[string]interface{}, error) {
	// Es wird ermittelt ob es sich um zulässiges Protokoll handelt
	if command == "ping_address" {
		// Es wird geprüft ob mindesten 1 Argument vorhanden ist
		if len(arguments) < 1 {
			return nil, fmt.Errorf("invalid ping command, has no arguments")
		}

		// Die Adresse wird versucht einzulesen
		pkey, err := btcec.ParsePubKey(arguments[0])
		if err != nil {
			return nil, fmt.Errorf("invalid public key")
		}

		// Der Pingbefehl wird verarbeitet und das Ergebniss wird zurückgegeben
		return obj._start_ping_pong_process(pkey)
	} else {
		return nil, fmt.Errorf("invalid command")
	}
}

// Registriert den Kernel im Protokoll
func (obj *ROUEX_PING_PONG_PROTOCOL) RegisterKernel(kernel *kernel.Kernel) error {
	obj._lock.Lock()
	if obj._kernel != nil {
		obj._lock.Unlock()
		return fmt.Errorf("kernel always registered")
	}
	obj._kernel = kernel
	obj._lock.Unlock()
	log.Println("ROUEX_PING_PONG_PROTOCOL: kernel registrated. id =", kernel.GetKernelID(), "object-id =", obj._objid)
	return nil
}

// Gibt den Namen des Protokolles zurück
func (obj *ROUEX_PING_PONG_PROTOCOL) GetProtocolName() string {
	return "ROUEX_PING_PONG_PROTOCOL"
}

// Gibt die ObjektID des Protokolls zurück
func (obj *ROUEX_PING_PONG_PROTOCOL) GetObjectId() string {
	return obj._objid
}

// Erzeugt ein neues PING PONG Protokoll
func NEW_ROUEX_PING_PONG_PROTOCOL_HANDLER() *ROUEX_PING_PONG_PROTOCOL {
	return &ROUEX_PING_PONG_PROTOCOL{_lock: &sync.Mutex{}, _objid: utils.RandStringRunes(12), _open_processes: []rouex_entry{}}
}
