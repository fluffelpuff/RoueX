package ipoverlay

import (
	"bytes"
	"encoding/hex"
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
	_otk_ecdh_key_id       string
	_is_connected          bool
	_ping_processes        []*PingProcess
	_conn                  *websocket.Conn
	_kernel                *kernel.Kernel
	_dest_relay_public_key *btcec.PublicKey
	_lock                  *sync.Mutex
	_write_lock            *sync.Mutex
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
	log.Println("WebsocketKernelConnection: kernel registered. connection =", obj._object_id, "kernel =", obj._kernel.GetKernelID())
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
	if len(obj._ping) >= 100 {
		obj._ping = append(obj._ping[:9], obj._ping[1:]...)
	}
	obj._lock.Unlock()
	log.Println("WebsocketKernelConnection: add ping time, connection =", obj._object_id, "time =", time)
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
	log.Println("WebsocketKernelConnection: new ping process created. connection =", new_proc.ObjectId, "process_id =", hex.EncodeToString(new_proc.ProcessId))
	return new_proc
}

// Wird verwendet um eine Pingsitzung zu entfernen
func (obj *WebsocketKernelConnection) _remove_ping_session(psession *PingProcess) {
	// Es wird geprüft ob die Verbindung vorhanden ist
	if !obj.IsConnected() {
		return
	}

	// Threadlock
	obj._lock.Lock()

	// Wird geprüft ob der Ping Prozess vorhanden ist
	is_found, hight := false, 0
	for i := range obj._ping_processes {
		if bytes.Equal(obj._ping_processes[i].ProcessId, psession.ProcessId) {
			is_found = true
			hight = i
			break
		}
	}

	// Sollte ein passender Eintrag gefunden wurden sein, wird dieser Entfernt
	if is_found {
		obj._ping_processes = append(obj._ping_processes[:hight], obj._ping_processes[hight+1:]...)
		log.Println("WebsocketKernelConnection: ping process removed. connection =", psession.ObjectId, "process_id =", hex.EncodeToString(psession.ProcessId))
	} else {
		log.Println("WebsocketKernelConnection: no ping process found to removing. connection =", psession.ObjectId, "process_id =", hex.EncodeToString(psession.ProcessId))
	}

	// Threadlock freigabe
	obj._lock.Unlock()
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
	log.Println("WebsocketKernelConnection: send ping package. connection =", obj._object_id, "process_id =", hex.EncodeToString(new_ping_session.ProcessId))
	if err := obj._write_ws_package(package_bytes, Ping); err != nil {
		return 0, err
	}

	// Es wird auf die Antwort des Paketes gewartet
	r_time, err := new_ping_session.untilWaitOfPong()
	if err != nil {
		return 0, err
	}

	// Log
	log.Println("WebsocketKernelConnection: ping for connection =", obj._object_id, "process_id =", hex.EncodeToString(new_ping_session.ProcessId), "time =", r_time)

	// Der Ping Vorgang wird wieder entfernt
	obj._remove_ping_session(new_ping_session)

	// Der Vorgang wurde ohne Fehler durchgeführt
	return r_time, nil
}

// Wird als eigenständiger Thread ausgeführt, sollte es sich um den ersten Erfolgreichen Ping vorgang handeln
func (obj *WebsocketKernelConnection) __first_ping_io_activated_routes_by_relay_connection() {
	// Es wird dem Kernel signalisiert dass alle bekannten Routen für die Relay Verbindung geladen werden sollen
	log.Println("WebsocketKernelConnection: try to load routes by relay connection. connection =", obj._object_id)
	obj._kernel.DumpsRoutesForRelayByConnection(obj)
}

// Gibt an ob der Ping Pong test korrekt ist
func (obj *WebsocketKernelConnection) __ping_auto_thread_pong() {
	// Speichert die Zeit ab, wann der Letze Ping durchgeführt wurde
	last_ping := time.Now()

	// Gibt an ob es sich um den ersten Start handelt
	is_first := true

	// Log
	log.Println("WebsocketKernelConnection: connection ping pong thread started. connection =", obj._object_id)

	// Wird solange ausgeführt, solange die Verbindung verbunden ist
	for obj.IsConnected() {
		// Zeitdauer seit last_ping messen
		elapsed := time.Since(last_ping)

		// Überprüfen, ob 10 Sekunden vergangen sind
		if elapsed.Seconds() >= 10 || (elapsed.Seconds() >= 1 && is_first) {
			// Es wird geprüft ob eine Verbindung mit der gegenseite besteht, wenn nicht wird der Vorgang abgebrochen
			if !obj.IsConnected() {
				break
			}

			// Es wird ein Ping vorgang durchgeführt
			w_time, err := obj._send_ping_and_wait_of_pong()
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Die Pingzeit wird abgespeichert
			obj._add_ping_time(w_time)

			// Es wird geprüft ob es sich um den ersten Ping vorgang handelt
			if is_first {
				// Es wird signalisiert dass alle Routen welche für diesen Relay verfügabr sind, geladen werden sollen
				go obj.__first_ping_io_activated_routes_by_relay_connection()

				// Es wird signalisiert dass es sich nicht mehr um den ersten Vorgang handelt
				is_first = false
			}

			// Speichert die Zeit des Pings ab
			last_ping = time.Now()
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}

	// Log
	log.Println("WebsocketKernelConnection: connection ping pong thread stoped. connection =", obj._object_id)
}

// Sendet dem Client ein Pong Paket zu
func (obj *WebsocketKernelConnection) __send_pong(ping_id []byte) error {
	// Das Pong Paket wird erzeugt
	pong_package := PongPackage{PingId: ping_id}

	// Das Paket wird in Bytes umgewandelt
	pong_package_bytes, err := pong_package.toBytes()
	if err != nil {
		panic(err)
	}

	// Es wird geprüft ob eine Verbindung mit der gegenseite besteht
	if !obj.IsConnected() {
		return fmt.Errorf("WebsocketKernelConnection.__send_pong: no connection")
	}

	// Das Paket wird versendet
	send_err := obj._write_ws_package(pong_package_bytes, Pong)
	if err != nil {
		return fmt.Errorf("__send_pong:" + send_err.Error())
	}

	// Der Vorgang wurde ohne fehler durchgeführt
	log.Println("WebsocketKernelConnection: pong package send. pingid =", hex.EncodeToString(ping_id))
	return nil
}

// Wird aufgerufen sobald ein Ping Paket eingetroffen ist
func (obj *WebsocketKernelConnection) __recived_ping_paket(data []byte) {
	// Es wird versucht das Ping Paket einzulesen
	ping_package, err := readPingPackageFromBytes(data)
	if err != nil {
		log.Println("Pong package sening error. pingid =", hex.EncodeToString(ping_package.PingId))
	}

	// Es wird versucht das Pong Paket zu senden
	if err := obj.__send_pong(ping_package.PingId); err != nil {
		if obj.IsConnected() {
			log.Println("__recived_ping_paket:" + err.Error())
		}
	}
}

// Wird aufgerufen sobald ein Pong Paket eingetroffen ist
func (obj *WebsocketKernelConnection) __recived_pong_paket(data []byte) {
	// Es wird versucht das Ping Paket einzulesen
	pong_package, err := readPongPackageFromBytes(data)
	if err != nil {
		log.Println("WebsocketKernelConnection: pong package reading error, aborted.")
		return
	}

	// Es wird geprüft ob ein Offener Ping Prozess mit dieser ID vorhanden ist
	obj._lock.Lock()
	var result *PingProcess
	for i := range obj._ping_processes {
		if bytes.Equal(obj._ping_processes[i].ProcessId, pong_package.PingId) {
			result = obj._ping_processes[i]
			break
		}
	}
	obj._lock.Unlock()

	// Sollte ein Prozess vorhanden sein, wird diesem Signalisiert dass eine Antwort empfangen wurde
	if result != nil {
		fmt.Println("WebsocketKernelConnection: pong recived", hex.EncodeToString(pong_package.PingId))
		result._signal_recived_pong()
	} else {
		fmt.Println("WebsocketKernelConnection: pong recived, unkown pong process", hex.EncodeToString(pong_package.PingId))
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
	obj._is_connected = true
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

		// Überprüfen Sie, ob der Nachrichtentyp "binary" ist
		if messageType != websocket.BinaryMessage {
			continue
		}

		// Das Paket wird anahnd des OTK Schlüssels entschlüsselt
		decrypted_package, err := obj._kernel.DecryptOTKECDHById(kernel.CHACHA_2020, obj._otk_ecdh_key_id, message)
		if err != nil {
			func_muutx.Lock()
			has_closed_reader_loop = err
			func_muutx.Unlock()
			break
		}

		// Das Verschlüsselte Paket wird eingelesen
		readed_ws_transport_paket, err := readWSTransportPaketFromBytes(decrypted_package)
		if err != nil {
			func_muutx.Lock()
			has_closed_reader_loop = err
			func_muutx.Unlock()
			break
		}

		// Die Signatur wird geprüft
		is_verify, err := kernel.VerifyByBytes(obj._dest_relay_public_key, readed_ws_transport_paket.Signature, readed_ws_transport_paket.Body)
		if err != nil {
			func_muutx.Lock()
			has_closed_reader_loop = err
			func_muutx.Unlock()
			break
		}
		if !is_verify {
			func_muutx.Lock()
			has_closed_reader_loop = fmt.Errorf("invalid package signature")
			func_muutx.Unlock()
			break
		}

		// Das Transportpaket wird entschlüsselt
		decrypted_transport_package, err := obj._kernel.DecryptWithPrivateRelayKey(readed_ws_transport_paket.Body)
		if err != nil {
			func_muutx.Lock()
			has_closed_reader_loop = err
			func_muutx.Unlock()
			break
		}

		// Es wird veruscht das Paket einzulesen
		read_transport_package, err := readEncryptedTransportPackageFromBytes(decrypted_transport_package)
		if err != nil {
			func_muutx.Lock()
			has_closed_reader_loop = err
			func_muutx.Unlock()
			break
		}

		// Es wird geprüft um was für ein Pakettypen es sich handelt
		if read_transport_package.Type == Ping {
			obj.__recived_ping_paket(read_transport_package.Data)
		} else if read_transport_package.Type == Pong {
			obj.__recived_pong_paket(read_transport_package.Data)
		} else {
			func_muutx.Lock()
			has_closed_reader_loop = fmt.Errorf("unkown package type")
			func_muutx.Unlock()
			break
		}

		// Es wird geprüft ob bereits ein Paket empfangen wurde
		func_muutx.Lock()
		if !has_recived {
			has_recived = true
		}
		func_muutx.Unlock()
	}

	// Es ein Fehler Vorhanden sein wird dieser angzeiegt
	if has_closed_reader_loop != nil {
		log.Println("WebsocketKernelConnection:", has_closed_reader_loop)
	}

	// Es wird signalisiert dass keine Verbindung mehr verfügbar ist
	obj._lock.Lock()
	obj._total_reader_threads--
	obj._is_connected = false
	obj._lock.Unlock()

	// Es wird geprüft ob das Objekt bereits fertigestellt wurde, wenn ja wird dem Kernel Signalisiert dass die Verbindung nicht mehr verfügbar ist
	obj._lock.Lock()
	if obj._is_finally {
		// Der Verbindung wird geschlossen
		_ = obj._conn.Close()
		obj._lock.Unlock()

		// Der Verbindung wird entfernt
		obj._kernel.RemoveConnection(nil, obj)
		return
	}
	obj._lock.Unlock()
}

// Wird verwendet um ein Paket Abzusenden
func (obj *WebsocketKernelConnection) _write_ws_package(data []byte, tpe TransportPackageType) error {
	// Das Transportpaket wird vorbereitet
	transport_package := EncryptedTransportPackage{SourceRelay: obj._kernel.GetPublicKey().SerializeCompressed(), DestinationRelay: obj._dest_relay_public_key.SerializeCompressed(), Type: tpe, Data: data}

	// Das Paket wird in Bytes umgewandelt
	byted_transport_package, err := transport_package.toBytes()
	if err != nil {
		panic(fmt.Errorf("_write_ws_package: " + err.Error()))
	}

	// Das Paket wird verschlüsselt
	encrypted_transport_package, err := kernel.EncryptECIESPublicKey(obj._dest_relay_public_key, byted_transport_package)
	if err != nil {
		return err
	}

	// Das Verschlüsselte Paket wird signiert
	sig, err := obj._kernel.SignWithRelayKey(encrypted_transport_package)
	if err != nil {
		panic(fmt.Errorf("_write_ws_package: " + err.Error()))
	}

	// Das Zwischenpaket wird erstellt
	twerp := WSTransportPaket{Body: encrypted_transport_package, Signature: sig}

	// Das Zwischenpaket wird in Bytes umgewandelt
	byted_twerp, err := twerp.toBytes()
	if err != nil {
		panic(fmt.Errorf("_write_ws_package: " + err.Error()))
	}

	// Das Paket wird mittels AES-256 Bit und dem OTK ECDH Schlüssel verschlüsselt
	final_encrypted, err := obj._kernel.EncryptOTKECDHById(kernel.CHACHA_2020, obj._otk_ecdh_key_id, byted_twerp)
	if err != nil {
		return err
	}

	// Das fertige Paket wird übertragen, hierfür wird der IO-Lock verwendet
	obj._write_lock.Lock()
	if err := obj._conn.WriteMessage(websocket.BinaryMessage, final_encrypted); err != nil {
		obj._write_lock.Unlock()
		return err
	}
	obj._write_lock.Unlock()

	// Der Vorgang wurde ohne Fehler erfolgreich druchgeführt
	log.Println("WebsocketKernelConnection: bytes writed. connection =", obj._object_id, "size =", len(final_encrypted))
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
	log.Println("WebsocketKernelConnection: finally connection. connection =", obj._object_id)
	return nil
}

// Gibt an ob eine Verbindung aufgebaut wurde
func (obj *WebsocketKernelConnection) IsConnected() bool {
	obj._lock.Lock()
	r := obj._is_connected
	t := obj._total_reader_threads
	s := obj._signal_shutdown
	obj._lock.Unlock()
	return r && t == 1 && !s
}

// Schreibt Daten in die Verbindung
func (obj *WebsocketKernelConnection) Write(data []byte) error {
	return obj._write_ws_package(data, Data)
}

// Gibt die Aktuelle Objekt ID aus
func (obj *WebsocketKernelConnection) GetObjectId() string {
	return obj._object_id
}

// Erstellt ein neues Kernel Sitzungs Objekt
func createFinallyKernelConnection(conn *websocket.Conn, local_otk_key_pair_id string, relay_public_key *btcec.PublicKey, relay_otk_public_key *btcec.PublicKey, relay_otk_ecdh_key_id string, bandwith float64, ping_time uint64) (*WebsocketKernelConnection, error) {
	// Das Objekt wird erstellt
	wkcobj := &WebsocketKernelConnection{
		_object_id:             kernel.RandStringRunes(12),
		_local_otk_key_pair:    local_otk_key_pair_id,
		_dest_relay_public_key: relay_public_key,
		_write_lock:            new(sync.Mutex),
		_lock:                  new(sync.Mutex),
		_otk_ecdh_key_id:       relay_otk_ecdh_key_id,
		_ping:                  []uint64{ping_time},
		_bandwith:              []float64{bandwith},
		_conn:                  conn,
		_signal_shutdown:       false,
	}

	// Das Objekt wird zurückgegben
	return wkcobj, nil
}
