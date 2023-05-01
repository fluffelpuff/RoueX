package ipoverlay

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/gorilla/websocket"
)

// Stellt eine Websocket Verbindung dar
type WebsocketKernelConnection struct {
	_object_id             string
	_local_otk_key_pair    string
	_total_reader_threads  uint8
	_is_finally            bool
	_signal_shutdown       bool
	_ping                  []uint64
	_bandwith              []float64
	_ping_processes        []*PingProcess
	_conn                  *websocket.Conn
	_kernel                *kernel.Kernel
	_dest_relay_public_key *btcec.PublicKey
	_lock                  *sync.Mutex
}

// Registriert einen Kernel in der Verbindung
func (obj *WebsocketKernelConnection) RegisterKernel(kernel *kernel.Kernel) error {
	obj._lock.Lock()
	if obj._kernel != nil {
		obj._lock.Unlock()
		return fmt.Errorf("kernel always registrated")
	}
	obj._kernel = kernel
	obj._lock.Unlock()
	return nil
}

// Gibt an ob die Lesende Schleife ausgeführt werden kann
func (obj *WebsocketKernelConnection) _loop_bckg_run() bool {
	obj._lock.Lock()
	r := obj._signal_shutdown
	obj._lock.Unlock()
	return !r
}

// Wird ausgeführt um eine neue Pingzeit hinzuzufügen
func (obj *WebsocketKernelConnection) _add_ping_time(time uint64) {
	obj._lock.Lock()
	obj._ping = append(obj._ping, time)
	obj._lock.Unlock()
}

// Wird ausgeführt um eine neuen Ping Vorgang zu Registrieren
func (obj *WebsocketKernelConnection) _create_new_ping_session() *PingProcess {
	// Es wird ein neuer Ping Vorgang registriert
	new_proc, err := newPingProcess()
	if err != nil {
		panic(err)
	}

	// Der Vorgang wird registriert
	obj._lock.Lock()
	obj._ping_processes = append(obj._ping_processes, new_proc)
	obj._lock.Unlock()

	// Die Daten werden zurückgegeben
	return new_proc
}

// Wird verwendet um ein Ping abzusenden und auf das Pong zu warten
func (obj *WebsocketKernelConnection) _send_ping_and_wait_of_pong() (uint64, error) {
	// Es wird ein neuer Ping vorgang registriert
	new_ping_session := obj._create_new_ping_session()

	// Das Ping Paket wird erzeugt
	ping_package, creating_error := new_ping_session.GetPingPackage()
	if creating_error != nil {
		panic(creating_error)
	}

	// Das Paket wird in Bytes umgewandelt
	package_bytes, err := ping_package.toBytes()
	if err != nil {
		panic(err)
	}

	// Das Paket wird gesendet
	obj._write_ws_package(package_bytes, Ping)

	// Es wird auf die Antwort des Paketes gewartet
	r_time, err := new_ping_session.untilWaitOfPong()
	if err != nil {
		return 0, err
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return r_time, nil
}

// Gibt an ob der Ping Pong test korrekt ist
func (obj *WebsocketKernelConnection) __ping_auto_thread_pong() {
	// Speichert die Zeit ab, wann der Letze Ping durchgeführt wurde
	last_ping := time.Now()

	// Wird solange ausgeführt, solange die Verbindung verbunden ist
	for obj.IsConnected() {
		// Zeitdauer seit last_ping messen
		elapsed := time.Since(last_ping)

		// Überprüfen, ob 10 Sekunden vergangen sind
		if elapsed.Seconds() >= 10 {
			// Es wird ein Ping vorgang durchgeführt
			w_time, err := obj._send_ping_and_wait_of_pong()
			if err != nil {
				break
			}

			// Die Pingzeit wird abgespeichert
			obj._add_ping_time(w_time)

			// Speichert die Zeit des Pings ab
			last_ping = time.Now()
		}
	}
}

// Nimmt eintreffende Pakete entgegen
func (obj *WebsocketKernelConnection) _thread_reader(rewolf chan string) {
	// Erzeugt den Funktions basierten Mutex
	func_muutx := sync.Mutex{}

	// Gibt an ob die Lesende Schleife beendet wurde
	var has_closed_reader_loop error

	// Gibt an ob bereits Daten empfangen wurden
	has_recived := false

	// Der Thread Signalisiert dass er ausgeführt wird
	obj._lock.Lock()
	obj._total_reader_threads++
	obj._lock.Unlock()

	// Diese Funktion wird nachdem Start ausgeführt, sie prüft 50 MS ob die Verbindung weiterhin besteht
	go func() {
		tick := 0
		for tick < 50 {
			func_muutx.Lock()
			if has_closed_reader_loop != nil {
				rewolf <- has_closed_reader_loop.Error()
				func_muutx.Unlock()
				return
			}
			if has_recived {
				rewolf <- "ok"
				func_muutx.Unlock()
				return
			}
			func_muutx.Unlock()
			time.Sleep(1 * time.Millisecond)
			tick++
		}
		rewolf <- "ok"
	}()

	// Diese Schleife wird solange ausgeführt bis die Verbindung getrennt / geschlossen wurde
	for obj._loop_bckg_run() {
		// Es wird auf eintreffende Pakete gewartet
		messageType, message, err := obj._conn.ReadMessage()
		if err != nil {
			func_muutx.Lock()
			has_closed_reader_loop = err
			func_muutx.Unlock()
			break
		}

		// Es wird geprüft ob bereits ein Paket empfangen wurde
		func_muutx.Lock()
		if !has_recived {
			has_recived = true
		}
		func_muutx.Unlock()

		// Überprüfen Sie, ob der Nachrichtentyp "binary" ist
		if messageType != websocket.BinaryMessage {
			continue
		}

		// Es wird versucht das Paket einzulesen
		fmt.Println(message)
	}

	// Es wird geprüft ob das Objekt bereits fertigestellt wurde, wenn ja wird dem Kernel Signalisiert dass die Verbindung nicht mehr verfügbar ist
	obj._lock.Lock()
	obj._total_reader_threads--
	if obj._is_finally {
		obj._lock.Unlock()
		obj._kernel.RemoveConnection(nil, obj)
		return
	}
	obj._lock.Unlock()
}

// Wird verwendet um ein Paket Abzusenden
func (obj *WebsocketKernelConnection) _write_ws_package(data []byte, tpe TransportPackageType) error {
	return nil
}

// Stellt die Verbindung vollständig fertig
func (obj *WebsocketKernelConnection) FinallyInit() error {
	// Es wird geprüft ob bereits ein Reader gestartet wurde
	obj._lock.Lock()
	if obj._total_reader_threads != 0 {
		obj._lock.Unlock()
		return nil
	}
	obj._lock.Unlock()

	// Der Reader wird gestartet
	io := make(chan string)
	go obj._thread_reader(io)

	// Es wird auf die Bestätigung durch den Reader gewartet
	resolv := <-io
	if resolv != "ok" {
		return fmt.Errorf(resolv)
	}

	// Es wird geprüft ob genau 1 Reader Thread ausgeführt wird
	obj._lock.Lock()
	if obj._total_reader_threads != 1 {
		obj._lock.Unlock()
		return fmt.Errorf("internal error")
	}

	// Es wird Signalisiert dass das Ojekt vollständig Finallisiert wurde
	obj._is_finally = true
	obj._lock.Unlock()

	// Der Ping Bandwith Thread wird gestartet, dieser ermittelt die Bandbreite sowie den Ping für diese Verbindung
	go obj.__ping_auto_thread_pong()

	// Der Vorgang wurde ohne einen Fehler durchgeführt
	log.Println("Finally connection", obj._object_id)
	return nil
}

// Gibt an ob eine Verbindung aufgebaut wurde
func (obj *WebsocketKernelConnection) IsConnected() bool {
	return true
}

// Schreibt Daten in die Verbindung
func (obj *WebsocketKernelConnection) Write(data []byte) error {
	return nil
}

// Gibt die Aktuelle Objekt ID aus
func (obj *WebsocketKernelConnection) GetObjectId() string {
	return obj._object_id
}

// Erstellt ein neues Kernel Sitzungs Objekt
func createFinallyKernelConnection(conn *websocket.Conn, local_otk_key_pair_id string, relay_public_key *btcec.PublicKey, relay_otk_public_key *btcec.PublicKey, bandwith float64, ping_time uint64) (*WebsocketKernelConnection, error) {
	wkcobj := &WebsocketKernelConnection{
		_object_id:             kernel.RandStringRunes(12),
		_local_otk_key_pair:    local_otk_key_pair_id,
		_dest_relay_public_key: relay_public_key,
		_lock:                  new(sync.Mutex),
		_signal_shutdown:       false,
		_ping:                  []uint64{ping_time},
		_bandwith:              []float64{bandwith},
		_conn:                  conn,
	}
	return wkcobj, nil
}
