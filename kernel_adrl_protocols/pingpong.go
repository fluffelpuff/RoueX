package kernelprotocols

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	addresspackages "github.com/fluffelpuff/RoueX/address_packages"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fluffelpuff/RoueX/kernel/extra"
	"github.com/fluffelpuff/RoueX/utils"
	"github.com/fxamacker/cbor"
)

// Stellt ein Ping Paket dar
type PingPongPackage struct {
	Type uint8
	Id   string
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
	_max_wait_time uint64
	_aborted       bool
	_wait_chan     chan bool
	_start_time    time.Time
	_finally_time  *time.Time
	_kernel        *kernel.Kernel
	_lock          *sync.Mutex
	_objid         string
}

// Gibt an ob die Maxiamle Zeit erreicht wurde
func (obj *rouex_entry) maxwnend() bool {
	// Die Zeitwerte werden abgerufen
	obj._lock.Lock()
	start_time := obj._start_time
	max_wait := obj._max_wait_time
	obj._lock.Unlock()

	// Der Wert wird zurückgegeben
	return time.Since(start_time) >= time.Duration(max_wait)*time.Second
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

	// Verwende ein select-Statement, um entweder den Channel-Empfang oder ein Timeout-Ereignis zu behandeln
	select {
	case msg := <-obj._wait_chan:
		// Es wird geprüft ob der Vorgang abgebrochen wurde
		if obj.isaborted() {
			return ABORTED, nil
		}

		// Es wird geprüft ob die Nachricht beantwortet wurde
		if msg {
			// Es wird zurückgegeben ob die Daten empfangen wurden
			return RESPONDED, nil
		} else {
			// Es wird geprüft ob der Vorgang abgebrochen wurde
			if obj.isaborted() {
				return ABORTED, nil
			}

			// Es wird signalisiert der Vorgang abgebrochen wurde
			obj._lock.Lock()
			obj._aborted = true
			obj._lock.Unlock()

			// Es wird Signalisiert dass der Vorgagn abgebrochen wurde
			return ABORTED, nil
		}
	case <-time.After(time.Duration(obj._max_wait_time) * time.Millisecond):
		// Es wird geprüft ob der Vorgang abgebrochen wurde
		if obj.isaborted() {
			return ABORTED, nil
		}

		// Es wird Siginalsiert dass der Vorgang abegrochen wurde
		obj._lock.Lock()
		obj._aborted = true
		obj._lock.Unlock()

		// Es wird Signalisiert dass es sich um ein Timeout handelt
		return TIMEOUT, nil
	}
}

// Gibt die Benötigte Zeit an
func (obj *rouex_entry) gtimems() uint64 {
	// Die Daten werden mit dem Threadlock abgerufen wird
	obj._lock.Lock()
	stime := obj._start_time
	etime := obj._finally_time
	obj._lock.Unlock()

	// Sollte noch kein Endzeit vorhanden sein wird der Vorgang abgebrochen
	if etime == nil {
		return math.MaxUint32
	}

	// Die benötigte Zeit wird ausgerechnet
	total_time := etime.Sub(stime)

	// Der Rückgabewert wird erstellt
	reval := uint64(total_time.Milliseconds())
	if reval < 1 {
		reval = 1
	}

	// Die benötigte Zeit wird zurückgegeben
	return reval
}

// Wird verwendet um zu Signalisieren dass eine Antwort eingetroffen ist
func (obj *rouex_entry) signal_response() {
	// Der Threadlock wird verwendet
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob der Aktuelle Stauts gesetzt wurde
	if obj._finally_time != nil {
		return
	}

	// Die Aktuelle Zeit wird gesetzt
	tr := time.Now()
	obj._finally_time = &tr

	// Dem Chan wird Signalisiert dass das Paket eingetroffen ist
	select {
	case obj._wait_chan <- true:
	default:
	}
}

// Signalisiert dass der Vorgang geschlossen wurde
func (obj *rouex_entry) Close() {
	// Der Threadlock wird verwendet
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird Signalisiert dass der Vorgang abgebrochen wurde
	obj._aborted = true

	// Dem Chan wird Signalisiert dass kein Paket eingetroffen ist
	select {
	case obj._wait_chan <- false:
	default:
	}
}

// Gibt die ID des Objektes zurück
func (obj *rouex_entry) GetId() string {
	return obj._id
}

// Stellt das Ping Protokoll dar
type ROUEX_PING_PONG_PROTOCOL struct {
	_open_processes map[string]*rouex_entry
	_objid          string
	_kernel         *kernel.Kernel
	_lock           *sync.Mutex
}

// Fügt einen Ping Prozess hinzu
func (obj *ROUEX_PING_PONG_PROTOCOL) _add_ping_process(ping_proc *rouex_entry, process_api_conn *kernel.APIProcessConnectionWrapper) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Speichert den Vorgang ab
	obj._open_processes[ping_proc._id] = ping_proc

	// Der Threadlock wird freigegebeb
	obj._lock.Unlock()

	// Die Verbindung wird gloabl gespeichert
	if process_api_conn != nil {
		process_api_conn.AddProcessInvigoratingService(ping_proc)
	}

	// Log
	log.Printf("ROUEX_PING_PONG_PROTOCOL: register new ping process. pid = %s\n", ping_proc.GetId())

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

// Entfernt einen Ping Prozess
func (obj *ROUEX_PING_PONG_PROTOCOL) _remove_ping_process(ping_proc *rouex_entry, process_api_conn *kernel.APIProcessConnectionWrapper) {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Der Ping Vorgang wird ermittelt
	_, found := obj._open_processes[ping_proc._id]
	if !found {
		return
	}

	// Der Eintrag wird entfernt
	delete(obj._open_processes, ping_proc._id)

	// Der Threadlock wird freigegebeb
	obj._lock.Unlock()

	// Die Verbindung wird gloabl entfernt
	if process_api_conn != nil {
		process_api_conn.RemoveProcessInvigoratingService(ping_proc)
	}

	// Log
	log.Printf("ROUEX_PING_PONG_PROTOCOL: ping process removed. pid = %s\n", ping_proc.GetId())
}

// Führt einen Ping Prozess durch
func (obj *ROUEX_PING_PONG_PROTOCOL) _start_ping_pong_process(pkey *btcec.PublicKey, process_api_conn *kernel.APIProcessConnectionWrapper) (map[string]interface{}, error) {
	// Die Prozess ID wird erstellt
	proc_id := utils.RandStringRunes(16)

	// Das Paket wird gebaut
	builded_package := PingPongPackage{Id: proc_id, Type: 0}

	// Das Paket wird in Bytes umgewandelt
	encoded_ping_package, err := cbor.Marshal(builded_package, cbor.EncOptions{})
	if err != nil {
		panic(err)
	}

	// Es wird ein neuer Ping Prozess wird erzeugt
	rx_entry := &rouex_entry{
		_id:            proc_id,
		_lock:          new(sync.Mutex),
		_kernel:        obj._kernel,
		_start_time:    time.Now(),
		_max_wait_time: uint64(1200),
		_objid:         utils.RandStringRunes(12),
		_wait_chan:     make(chan bool),
		_aborted:       false,
	}

	// Der Vorgang wird abgespeichert
	if err := obj._add_ping_process(rx_entry, process_api_conn); err != nil {
		return nil, fmt.Errorf("_start_ping_pong_process: " + err.Error())
	}

	// Das Ping Paket wird über das Netzwerk übermittelt
	sstate, err := obj._kernel.EnterBytesEncryptAndSendL2PackageToNetwork(0, encoded_ping_package, pkey)
	if err != nil {
		// Der Ping Prozess wird wieder entfernt
		obj._remove_ping_process(rx_entry, process_api_conn)

		// Der Fehler wird zurückgegeben
		return nil, fmt.Errorf("_start_ping_pong_process: " + err.Error())
	}

	// Das Rückgabeobjekt wird erstellt
	reval := make(map[string]interface{})

	// Es wird gewartet bis sich der Status des Paketes geändert hat
	for range time.Tick(1 * time.Millisecond) {
		if sstate.GetState() != extra.WAIT {
			break
		}
		if !obj._kernel.IsRunning() {
			break
		}
		if rx_entry.isaborted() {
			break
		}
	}

	// Es wird geprüft ob der Vorgang abgebrochen
	if rx_entry.isaborted() {
		log.Printf("ROUEX_PING_PONG_PROTOCOL: ping process aborted. pid = %s\n", proc_id)
		obj._remove_ping_process(rx_entry, process_api_conn)
		reval["state"] = uint8(ABORTED)
		return reval, nil
	}

	// Log
	log.Printf("ROUEX_PING_PONG_PROTOCOL: ping package transmitted. pid = %s\n", proc_id)

	// Es wird auf die Antwort wird gewartet
	state, err := rx_entry.waitfnc()
	if err != nil {
		// Der Pingvorgang wird entfernt
		obj._remove_ping_process(rx_entry, process_api_conn)

		// Der Fehler wird zurückgegeben
		return nil, fmt.Errorf("_start_ping_pong_process: " + err.Error())
	}

	// Es wird geprüft das der Vorgang mit einem Response beantwortet wurde
	switch state {
	case ABORTED:
		log.Printf("ROUEX_PING_PONG_PROTOCOL: ping process aborted. pid = %s\n", proc_id)
		obj._remove_ping_process(rx_entry, process_api_conn)
		reval["state"] = uint8(ABORTED)
		return reval, nil
	case RESPONDED:
		log.Printf("ROUEX_PING_PONG_PROTOCOL: ping process responded. pid = %s, total = %d ms\n", proc_id, rx_entry.gtimems())
		obj._remove_ping_process(rx_entry, process_api_conn)
		reval["ttime"] = rx_entry.gtimems()
		reval["state"] = uint8(RESPONDED)
		return reval, nil
	case CLOSED_BY_KERNEL:
		log.Printf("ROUEX_PING_PONG_PROTOCOL: ping process closed by kernel. pid = %s\n", proc_id)
		obj._remove_ping_process(rx_entry, process_api_conn)
		reval["state"] = uint8(CLOSED_BY_KERNEL)
		return reval, nil
	case TIMEOUT:
		log.Printf("ROUEX_PING_PONG_PROTOCOL: ping process time out. pid = %s\n", proc_id)
		obj._remove_ping_process(rx_entry, process_api_conn)
		reval["state"] = uint8(TIMEOUT)
		return reval, nil
	default:
		obj._remove_ping_process(rx_entry, process_api_conn)
		log.Println("Error by handling connection", state)
		return nil, fmt.Errorf("unkown state")
	}
}

// Nimmt eintreffende Ping Pakete engegeen
func (obj *ROUEX_PING_PONG_PROTOCOL) _enter_incomming_ping_package(ppp PingPongPackage, source *btcec.PublicKey) error {
	// Das Paket wird gebaut
	builded_package := PingPongPackage{Id: ppp.Id, Type: 1}

	// Das Paket wird in Bytes umgewandelt
	encoded_pong_package, err := cbor.Marshal(builded_package, cbor.EncOptions{})
	if err != nil {
		panic(err)
	}

	// Log
	log.Println("ROUEX_PING_PONG_PROTOCOL: ping package recived. id = "+ppp.Id, "source = "+hex.EncodeToString(source.SerializeCompressed()))

	// Das Ping Paket wird über das Netzwerk übermittelt
	_, err = obj._kernel.EnterBytesEncryptAndSendL2PackageToNetwork(0, encoded_pong_package, source)
	if err != nil {
		return fmt.Errorf("sending error: " + err.Error())
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

// Nimmt eintreffende Pong Pakete entgegen
func (obj *ROUEX_PING_PONG_PROTOCOL) _enter_incomming_pong_package(ppp PingPongPackage, source *btcec.PublicKey) error {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob es einen offnenen Vorgang gibt
	pro, no_found := obj._open_processes[ppp.Id]
	if !no_found {
		return nil
	}

	// Es wird geprüft ob das Paket beretis beantwortet wurde
	if pro.isfinn() {
		return nil
	}

	// Es wird Signalisiert dass ein Pong empangen wurde
	pro.signal_response()

	// Der Vorgang wurde ohen Fehler durchgeführt
	return nil
}

// Nimmt eingetroffene Pakete aus dem Netzwerk Entgegen
func (obj *ROUEX_PING_PONG_PROTOCOL) EnterRecivedPackage(pckage *addresspackages.AddressLayerPackage) error {
	// Es wird versucht das Paket einzulesen
	var ppp PingPongPackage
	if err := cbor.Unmarshal(pckage.Data, &ppp); err != nil {
		return fmt.Errorf("error: invalid_package: " + err.Error())
	}

	// Es wird geprüft ob es sich um ein Ping oder um ein Pong Paket handelt
	switch ppp.Type {
	case 0:
		return obj._enter_incomming_ping_package(ppp, &pckage.Sender)
	case 1:
		return obj._enter_incomming_pong_package(ppp, &pckage.Sender)
	default:
		return fmt.Errorf("error: invalid package type")
	}
}

// Nimmt eintreffende Steuer Befehele entgegen
func (obj *ROUEX_PING_PONG_PROTOCOL) EnterCommandData(command string, arguments [][]byte, process_api_conn *kernel.APIProcessConnectionWrapper) (map[string]interface{}, error) {
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
		return obj._start_ping_pong_process(pkey, process_api_conn)
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
	return &ROUEX_PING_PONG_PROTOCOL{_lock: &sync.Mutex{}, _objid: utils.RandStringRunes(12), _open_processes: make(map[string]*rouex_entry)}
}
