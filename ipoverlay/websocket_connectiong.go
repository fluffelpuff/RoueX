package ipoverlay

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	addresspackages "github.com/fluffelpuff/RoueX/address_packages"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fluffelpuff/RoueX/kernel/extra"
	"github.com/fluffelpuff/RoueX/static"
	"github.com/fluffelpuff/RoueX/utils"
	"github.com/gorilla/websocket"
)

// Stellt einen write_buffer eintrag dar
type writer_buffer_entry struct {
	data   []byte
	sstate *extra.PackageSendState
	size   uint64
}

// Stellt eine Websocket Verbindung dar
type WebsocketKernelConnection struct {
	_object_id               string
	_local_otk_key_pair      string
	_total_reader_threads    uint8
	_total_ping_pong_threads uint8
	_total_writer_threads    uint8
	_is_finally              bool
	_signal_shutdown         bool
	_disconnected            bool
	_ping                    []uint64
	_bandwith                []float64
	_otk_ecdh_key_id         string
	_is_connected            bool
	_destroyed               bool
	_ping_processes          []*PingProcess
	_conn                    *websocket.Conn
	_kernel                  *kernel.Kernel
	_dest_relay_public_key   *btcec.PublicKey
	_lock                    *sync.Mutex
	_write_lock              *sync.Mutex
	_io_type                 kernel.ConnectionIoType
	_rx_bytes                uint64
	_tx_bytes                uint64
	_write_buffer            []*writer_buffer_entry
	_writer_buffer_total     uint64
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

// Die Verbindung wurde geschlossen
func (obj *WebsocketKernelConnection) _destroy_disconnected() {
	obj._lock.Lock()
	if obj._is_finally {
		// Der Verbindung wird geschlossen
		_ = obj._conn.Close()

		// Es werden alle offenen Ping vorgänge zerstört
		for i := range obj._ping_processes {
			obj._ping_processes[i]._signal_abort()
		}

		// Der Threadlock wird freigegeben
		obj._lock.Unlock()

		// Der Verbindung wird entfernt
		if err := obj._kernel.RemoveConnection(obj); err != nil {
			log.Println("WebsocketKernelConnection: error by destorying conenction. error =", err.Error())
		}

		// Das Objekt wird als zerstört Markiert
		obj._lock.Lock()
		obj._destroyed = true
		obj._lock.Unlock()

		// Der Vorgang wird beenet
		return
	}
	obj._lock.Unlock()
}

// Gibt an ob die Lesende Schleife ausgeführt werden kann
func (obj *WebsocketKernelConnection) _loop_bckg_run() bool {
	obj._lock.Lock()
	r := obj._signal_shutdown
	t := obj._disconnected
	obj._lock.Unlock()
	return !r && !t
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

// Nimmt eintreffende Datenpakete entgegeen
func (obj *WebsocketKernelConnection) _enter_incomming_data_package(data []byte) {
	// Es wird geprüft ob das Paket größer als 30 Bytes ist
	if len(data) < 30 {
		log.Println("WebsocketKernelConnection: invalid data package recived. connection =", obj._object_id)
	}

	// Es wird versucht das Package Frame einzulesen
	readed_package, err := addresspackages.ReadFinalAddressLayerPackageFromBytes(data)
	if err != nil {
		log.Println("WebsocketKernelConnection: error by reading, package droped. error = "+err.Error(), "connection = "+obj._object_id)
		return
	}

	// Es wird geprüft ob die Signatur korrekt ist
	if !readed_package.ValidateSignature() {
		log.Println("WebsocketKernelConnection: error by reading, package droped. error = invalid package siganture. connection = " + obj._object_id)
	}

	// Das Paket wird an den Kernel übergeben
	if err := obj._kernel.EnterL2Package(readed_package, obj); err != nil {
		log.Println(err)
		return
	}
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
	if complete, was_added := obj._kernel.DumpsRoutesForRelayByConnection(obj); complete && was_added {
		log.Println("WebsocketKernelConnection: routes initing completed. connection =", obj._object_id)
	}
}

// Gibt an ob der Ping Pong test korrekt ist
func (obj *WebsocketKernelConnection) __ping_auto_thread_pong() {
	// Speichert die Zeit ab, wann der Letze Ping durchgeführt wurde
	last_ping := time.Now()

	// Gibt an ob es sich um den ersten Start handelt
	is_first := true

	// Log
	log.Println("WebsocketKernelConnection: connection ping pong thread started. connection =", obj._object_id)

	// Signalisiert dass ein Ping Pong Thread vorhanden ist
	obj._lock.Lock()
	obj._total_ping_pong_threads++
	obj._lock.Unlock()

	// Wird solange ausgeführt, solange die Verbindung verbunden ist
	for obj._loop_bckg_run() {
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
				log.Println(err)
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

	// Signalisiert dass ein Ping Pong Thread geschlossen wurde
	obj._lock.Lock()
	obj._total_ping_pong_threads--
	obj._lock.Unlock()

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
		log.Println("WebsocketKernelConnection: pong recived", hex.EncodeToString(pong_package.PingId))
		result._signal_recived_pong()
	} else {
		log.Println("WebsocketKernelConnection: pong recived, unkown pong process", hex.EncodeToString(pong_package.PingId))
	}
}

// Gibt an, wieviele Ping Ping Processes vorhanden sind
func (obj *WebsocketKernelConnection) _ping_ponger_is_open() bool {
	obj._lock.Lock()
	r := obj._total_ping_pong_threads
	obj._lock.Unlock()
	return r > 0
}

// Signalisiert dass die Verbindung gerennt wurde
func (obj *WebsocketKernelConnection) _signal_disconnect() {
	// Signalisiert dass die Verbindung getrennt wurde
	obj._lock.Lock()
	obj._disconnected = true
	obj._lock.Unlock()

	// Es wird gewartet bis alle Ping Pong Threads geschlossen wurden
	for obj._ping_ponger_is_open() {
		time.Sleep(1 * time.Millisecond)
	}
}

// Wird verwendet um die eigentlichen Daten zu senden
func (obj *WebsocketKernelConnection) _write_operation() {
	// Diese Funktion gibt an ob Daten Verfügbar sind
	sendable_avail := func(c *WebsocketKernelConnection) bool {
		// Der Threadlock wird ausgeführt
		c._lock.Lock()
		defer c._lock.Unlock()

		// Der Aktuelle Status wird zurückgegeben
		return len(c._write_buffer) > 0
	}

	// Ruft das erste Paket ab und versendet es
	get_pckg := func(c *WebsocketKernelConnection) (*extra.PackageSendState, []byte, bool) {
		// Der Threadlock wird ausgeführt
		c._lock.Lock()
		defer c._lock.Unlock()

		// Es wird geprüft ob ein Paket verfügbar ist
		if len(c._write_buffer) == 0 {
			return nil, nil, false
		}

		// Das Paket wird aus dem Buffer abgerufen
		n_package := c._write_buffer[0]

		// Das Paket wird aus dem Buffer entfernt
		c._write_buffer = append(c._write_buffer[:0], c._write_buffer[1:]...)
		obj._writer_buffer_total -= uint64(len(n_package.data))

		// Die Daten werden zurückgegeben
		return n_package.sstate, n_package.data, true
	}

	// Diese Schleife wird solange ausgeführt bis keine Nachrichten mehr verfügbar sind
	for sendable_avail(obj) {
		// Das Aktuell zu sendene Paket wird abgerufen
		c_state, data, avail := get_pckg(obj)
		if !avail {
			break
		}

		// Die Daten werden gesendet
		if err := obj.Write(data); err != nil {
			// Es wird Signalisiert dass die Daten nicht gesendet werden konnten
			c_state.SetFinallyState(extra.DROPED)

			// Der Vorgang wird beendet
			continue
		}

		// Es wird Signalisiert dass das Paket erfolgreich gesendet wurde
		c_state.SetFinallyState(extra.SEND)

		// Es wird 100 Nanaosekunden gewartet
		time.Sleep(100 * time.Nanosecond)
	}
}

// Nimmt eintreffende Pakete entgegen
func (obj *WebsocketKernelConnection) _start_thread_reader() error {
	// Erzeugt den Funktions basierten Mutex
	func_muutx := sync.Mutex{}

	// Gibt an ob die Lesende Schleife beendet wurde
	var has_closed_reader_loop error

	// Der Reader Thread wird gestartet
	go func() {
		// Der Thread Signalisiert dass er ausgeführt wird
		obj._lock.Lock()
		obj._total_reader_threads++
		obj._is_connected = true
		obj._lock.Unlock()

		// Diese Schleife wird solange ausgeführt bis die Verbindung getrennt / geschlossen wurde
		for obj._loop_bckg_run() {
			// Es wird auf eintreffende Pakete gewartet
			messageType, message, err := obj._conn.ReadMessage()
			if err != nil {
				obj._lock.Lock()
				if obj._signal_shutdown {
					obj._lock.Unlock()
					break
				}
				obj._lock.Unlock()
				func_muutx.Lock()
				has_closed_reader_loop = err
				func_muutx.Unlock()
				break
			}

			// Die Bytes werden zugeordnert
			obj._lock.Lock()
			obj._rx_bytes += uint64(len(message))
			obj._lock.Unlock()

			// Überprüfen Sie, ob der Nachrichtentyp "binary" ist
			if messageType != websocket.BinaryMessage {
				continue
			}

			// Das Paket wird anahnd des OTK Schlüssels entschlüsselt
			decrypted_package, err := obj._kernel.DecryptOTKECDHById(utils.CHACHA_2020, obj._otk_ecdh_key_id, message)
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
			is_verify, err := utils.VerifyByBytes(obj._dest_relay_public_key, readed_ws_transport_paket.Signature, readed_ws_transport_paket.Body)
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
			} else if read_transport_package.Type == Data {
				obj._enter_incomming_data_package(read_transport_package.Data)
			} else {
				func_muutx.Lock()
				has_closed_reader_loop = fmt.Errorf("unkown package type")
				func_muutx.Unlock()
				break
			}
		}

		// Log
		log.Println("WebsocketKernelConnection: connection closed. connection =", obj._object_id)

		// Signalisiert dass die Verbindung getrennt wurde
		obj._signal_disconnect()

		// Es wird signalisiert dass keine Verbindung mehr verfügbar ist
		obj._lock.Lock()
		obj._total_reader_threads--
		obj._is_connected = false
		obj._lock.Unlock()

		// Die Verbindung wird vollständig vernichtet
		obj._destroy_disconnected()
	}()

	// Es wird auf die Bestätigung durch den Reader gewartet (20ms)
	for i := 1; i <= 20; i++ {
		time.Sleep(1 * time.Millisecond)
	}

	// Es wird geprüft ob ein Fehler aufgetreten ist
	func_muutx.Lock()
	defer func_muutx.Unlock()
	if has_closed_reader_loop != nil {
		return fmt.Errorf("_start_thread_reader: " + has_closed_reader_loop.Error())
	}

	// Es wird geprüft ob die Verbindung aufgebaut werden konnte
	if !obj.IsConnected() {
		return fmt.Errorf("_start_thread_reader: connection was closed by initing")
	}

	// Der Vorgang wurde Fehler durchgeführt
	return nil
}

// Wird als Thread ausgeführt und versendet ausgehende Pakete
func (obj *WebsocketKernelConnection) _start_thread_writer() error {
	// Gibt an ob der Writer gestartet wurde
	was_started := false

	// Der Writer wird als Thread ausgeführt
	go func() {
		// Es wird Signalisiert dass der Writer ausgeführt wird
		obj._lock.Lock()
		obj._total_writer_threads++
		was_started = true
		obj._lock.Unlock()

		// Die Schleife wird solange ausgeführt, bis die Verbindung getrennt wurde
		for obj.IsConnected() {
			// Die Operation wird druchgeführt
			obj._write_operation()

			// Es wird eine MS gewartet
			time.Sleep(1 * time.Millisecond)
		}

		// Es wird Signalisiert dass der Writer nicht mehr ausgeführt wird
		obj._lock.Lock()
		obj._total_writer_threads--
		obj._lock.Unlock()
	}()

	// Es wird gewartet bis der Writer gestartet wurde
	for {
		obj._lock.Lock()
		if was_started {
			obj._lock.Unlock()
			break
		}
		obj._lock.Unlock()
		time.Sleep(1 * time.Millisecond)
	}

	// Der Vorgang wurde ohne Fehler gestartet
	return nil
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
	encrypted_transport_package, err := utils.EncryptECIESPublicKey(obj._dest_relay_public_key, byted_transport_package)
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
	final_encrypted, err := obj._kernel.EncryptOTKECDHById(utils.CHACHA_2020, obj._otk_ecdh_key_id, byted_twerp)
	if err != nil {
		return err
	}

	// Der Threadlock wird verwendet
	obj._write_lock.Lock()

	// Das fertige Paket wird übertragen, hierfür wird der IO-Lock verwendet
	if err := obj._conn.WriteMessage(websocket.BinaryMessage, final_encrypted); err != nil {
		obj._write_lock.Unlock()
		return err
	}

	// Die gesendeten Bytes werden hinzugerechnet
	obj._tx_bytes += uint64(len(final_encrypted))

	// Der Threadlock wird freigegeben
	obj._write_lock.Unlock()

	// Der Vorgang wurde ohne Fehler erfolgreich druchgeführt
	log.Println("WebsocketKernelConnection: bytes writed. connection =", obj._object_id, "size =", len(final_encrypted))
	return nil
}

// Gibt an ob die Websocket Verbindung geschlossen wurde
func (obj *WebsocketKernelConnection) _check_is_full_closed() bool {
	// Der Threadlock wird verwendet
	obj._lock.Lock()

	// Es wird ermittelt ob der Socket verbunden ist
	r := !obj._is_connected

	// Es wird ermittelt wieviele Ping Prozesse vorhanden sind
	p := len(obj._ping_processes) == 0

	// Es wird ermittelt ob alle Reader geschlossen wurden
	t := obj._total_reader_threads == 0

	// Gibt an ob das Objekt zerstört wurde
	x := obj._destroyed

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Das Ergebniss wird zurückgegeben
	return r && p && t && x
}

// Stellt die Verbindung vollständig fertig
func (obj *WebsocketKernelConnection) FinallyInit() error {
	// Es wird geprüft ob bereits ein Reader gestartet wurde
	obj._lock.Lock()
	if obj._total_reader_threads != 0 || obj._total_writer_threads != 0 {
		obj._lock.Unlock()
		return nil
	}
	obj._lock.Unlock()

	// Der Reader wird gestartet
	if err := obj._start_thread_reader(); err != nil {
		return fmt.Errorf("FinallyInit: 1: " + err.Error())
	}

	// Der Senderthread wird gestartet
	if err := obj._start_thread_writer(); err != nil {
		return fmt.Errorf("FinallyInit: 2: " + err.Error())
	}

	// Es wird geprüft ob der Reader und Writer Thread bereits ausgeführt wird
	obj._lock.Lock()
	if obj._total_reader_threads != 1 || obj._total_writer_threads != 1 {
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

// Gibt das verwendete Protokoll an
func (obj *WebsocketKernelConnection) GetProtocol() string {
	return "ws"
}

// Gibt an ob eine Verbindung aufgebaut wurde
func (obj *WebsocketKernelConnection) IsConnected() bool {
	obj._lock.Lock()
	r := obj._is_connected
	t := obj._total_reader_threads
	s := obj._signal_shutdown
	x := obj._disconnected
	obj._lock.Unlock()
	return r && t == 1 && !s && !x
}

// Schreibt Daten in die Verbindung
func (obj *WebsocketKernelConnection) Write(data []byte) error {
	return obj._write_ws_package(data, Data)
}

// Gibt die Aktuelle Objekt ID aus
func (obj *WebsocketKernelConnection) GetObjectId() string {
	return obj._object_id
}

// Wird verwendet um die Verbindung zu schließen
func (obj *WebsocketKernelConnection) CloseByKernel() {
	// Der Threadlock wird verwendet
	obj._lock.Lock()

	// Es wird geprüft ob bereits Signalisiert wurde ob die Verbindung geschlossen werden soll
	if obj._signal_shutdown {
		obj._lock.Unlock()
		return
	}

	// Es wird geprüft ob die Verbindung bereits getrennt wurde
	if !obj._is_connected {
		obj._lock.Unlock()
		return
	}

	// Log
	log.Println("WebsocketKernelConnection: close connection by kernel. connection =", obj._object_id)

	// Es wird Signalisiert dass es sich um einen Schließvorgnag handelt
	obj._signal_shutdown = true

	// Die Werbsocket Verbindung wird geschlossen
	obj._conn.Close()

	// Der Threadlock wird freigegeben_signal_shutdown
	obj._lock.Unlock()

	// Wartet solange bis die Schleife geschlossen wurde
	for !obj._check_is_full_closed() {
		time.Sleep(1 * time.Millisecond)
	}
}

// Gibt die Aktuelle Ping Zeit zurück
func (obj *WebsocketKernelConnection) GetPingTime() uint64 {
	obj._lock.Lock()
	if len(obj._ping) == 0 {
		obj._lock.Unlock()
		return 0
	}
	r := obj._ping[len(obj._ping)-1]
	obj._lock.Unlock()
	return r
}

// Gibt die Gesendete und Empfangene Datenmenge zurück
func (obj *WebsocketKernelConnection) GetTxRxBytes() (uint64, uint64) {
	obj._lock.Lock()
	r, t := obj._rx_bytes, obj._tx_bytes
	obj._lock.Unlock()
	return r, t
}

// Gibt an ob es sich um eine ein oder ausgehende Verbindung handelt
func (obj *WebsocketKernelConnection) GetIOType() kernel.ConnectionIoType {
	// Der Threadlock wird verwendet
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Der Aktuelle Verbindungstyp wird zurückgegeben
	return obj._io_type
}

// Gibt den Öffentlichen Sitzungsschlüssel zurück
func (obj *WebsocketKernelConnection) GetSessionPKey() (*btcec.PublicKey, error) {
	r, err := obj._kernel.GetPublicTempKeyById(obj._local_otk_key_pair)
	return r, err
}

// Gibt an, ob es sich um die gleiche Verbindung handelt
func (obj *WebsocketKernelConnection) Equal(p2 *WebsocketKernelConnection) bool {
	return obj.GetObjectId() == p2.GetObjectId()
}

// Gibt den Hashwert des Objekts zurück
func (obj *WebsocketKernelConnection) Hash() uint32 {
	var hash uint32
	for _, c := range obj.GetObjectId() {
		hash = 31*hash + uint32(c)
	}
	return hash
}

// Gibt an ob die Verbindung fertigestellt wurde
func (obj *WebsocketKernelConnection) IsFinally() bool {
	obj._lock.Lock()
	defer obj._lock.Unlock()
	return obj._is_finally
}

// Wird verwendet um Gepufferte Daten entgegenzunehemen
func (obj *WebsocketKernelConnection) EnterSendableData(data []byte) (*extra.PackageSendState, error) {
	// Das Rückgabe Objekt wird gebaut
	revobj := extra.NewPackageSendState()

	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Es wird geprüft ob der Puffer über ausreichend Speicher verfügt
	if len(obj._write_buffer) >= int(static.WS_MAX_PACKAGES) || obj._writer_buffer_total+uint64(len(data)) > static.WS_MAX_BYTES {
		// Dem Eintrag wird Signalisiert dass die Daten verworfen wurden
		obj._write_buffer[0].sstate.SetFinallyState(extra.DROPED)

		// Der Eintrag wird aus dem Stack entfernt
		obj._writer_buffer_total -= uint64(len(obj._write_buffer[1].data))
		obj._write_buffer = append(obj._write_buffer[0:], obj._write_buffer[1:]...)
	}

	// Der Eintrag wird in dem Buffer zwischengespeichert
	obj._write_buffer = append(obj._write_buffer, &writer_buffer_entry{data: data, sstate: revobj, size: uint64(len(data))})
	obj._writer_buffer_total += uint64(len(data))

	// Der Vorgang wurde ohne Fehler erfolreich fertigestellt
	return revobj, nil
}

// Gibt an ob der Buffer der Verbindung das Schreiben zu lässt
func (obj *WebsocketKernelConnection) CannUseToWrite() bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Der Vorgang wurde ohne Fehler erfolreich fertigestellt
	return !(len(obj._write_buffer) >= 2)
}

// Erstellt ein neues Kernel Sitzungs Objekt
func createFinallyKernelConnection(conn *websocket.Conn, local_otk_key_pair_id string, relay_public_key *btcec.PublicKey, relay_otk_public_key *btcec.PublicKey, relay_otk_ecdh_key_id string, bandwith float64, ping_time uint64, io_type kernel.ConnectionIoType) (*WebsocketKernelConnection, error) {
	// Das Objekt wird erstellt
	wkcobj := &WebsocketKernelConnection{
		_object_id:             utils.RandStringRunes(12),
		_local_otk_key_pair:    local_otk_key_pair_id,
		_dest_relay_public_key: relay_public_key,
		_write_lock:            new(sync.Mutex),
		_lock:                  new(sync.Mutex),
		_otk_ecdh_key_id:       relay_otk_ecdh_key_id,
		_ping:                  []uint64{ping_time},
		_bandwith:              []float64{bandwith},
		_write_buffer:          make([]*writer_buffer_entry, 0),
		_io_type:               io_type,
		_conn:                  conn,
		_signal_shutdown:       false,
		_total_writer_threads:  0,
		_total_reader_threads:  0,
	}

	// Das Objekt wird zurückgegben
	return wkcobj, nil
}
