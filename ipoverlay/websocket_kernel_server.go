package ipoverlay

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fluffelpuff/RoueX/utils"
	"github.com/fxamacker/cbor"
	"github.com/gorilla/websocket"
)

// Stellt das Upgrader Objekt da
var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// Stellt den Websocket Server dar
type WebsocketKernelServerEP struct {
	_kernel          *kernel.Kernel
	_obj_id          string
	_shutdown_signal bool
	_is_running      bool
	_total_proc      uint8
	_tcp_server      *http.Server
	_lock            *sync.Mutex
	_port            int
}

// Gibt an ob der Server ausgeführt wird
func (obj *WebsocketKernelServerEP) _is_rn() bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()
	defer obj._lock.Unlock()

	// Sollte noch mehr als 1 Proc ausgeführt werden, wird ein True zurückgegebn
	if obj._total_proc > 0 {
		return true
	}

	// Der Aktuelle Running Status wird zurückgegeb
	return obj._is_running
}

// Registriert den Kernel im Module
func (obj *WebsocketKernelServerEP) RegisterKernel(k *kernel.Kernel) error {
	log.Printf("Websocket Server EndPoint registrated on kernel %s\n", k.GetKernelID())
	obj._kernel = k
	return nil
}

// Wird verwendet um den Serversocket herunterzufahren
func (obj *WebsocketKernelServerEP) Shutdown() {
	// Log
	log.Println("WebsocketKernelServerEP: shutingdown. id =", obj._obj_id)

	// Der Threadlock wird gesperrt
	obj._lock.Lock()

	// Es wird signalisiert dass der Server heruntergefahren werden soll
	obj._shutdown_signal = true

	// Schließt den TCP Server
	obj._tcp_server.Close()

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Es wird gewartet bis der Server beendet wurde
	for obj._is_rn() {
		time.Sleep(1 * time.Millisecond)
	}
}

// Wird verwendet um den TCP basierten Server zu Starten
func (obj *WebsocketKernelServerEP) _start_tcp_tcp_server() error {
	// Der Websocket TCP Server wird erstellt
	obj._tcp_server = &http.Server{
		Addr:    ":" + strconv.Itoa(obj._port),
		Handler: http.HandlerFunc(obj.upgradeHTTPConnAndRegister),
	}

	// Gibt an ob der eigentliche Server ausgeführt wird
	is_running := false

	// Der Server wurde gestartet
	go func() {
		// Es wird signalisiert das der Thrad ausgeführt wird
		obj._lock.Lock()

		// Es wird angegeben das der Server ausgeführt wird
		is_running = true

		// Es wird Signalisiert dass ein weiterer Server Thread ausgeführt wird
		obj._total_proc++

		// Der Thread wird freigegeben
		obj._lock.Unlock()

		// Der Log wird angezeigt
		log.Printf("WebsocketKernelServerEP: new tcp based server started. id = %s, endpoint = %s\n", obj._obj_id, "0.0.0.0:"+strconv.Itoa(obj._port))
		if err := obj._tcp_server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}

		// Es wird versucht den Threadlock zu sperren
		obj._lock.Lock()

		// Es wird Signalisiert dass der Server beendet wurde
		is_running = false

		// Es wird Signalisiert dass ein Protokoll Thread weniger läuft
		obj._total_proc--

		// Der Threadlock wird freigeben
		obj._lock.Unlock()

		// Log
		log.Printf("WebsocketKernelServerEP: closed. id = %s, endpoint = %s\n", obj._obj_id, "0.0.0.0:"+strconv.Itoa(obj._port))
	}()

	// Diese Funktion gibt an dass der Server noch läuft
	server_running := func() bool {
		// Der Threadlock wird verwendet
		obj._lock.Lock()
		defer obj._lock.Unlock()

		// Der Status wird zurückgegeben
		return is_running
	}

	// Es wird 10 MS gewartet ob der Server noch ausgeführt wird
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Millisecond)
	}

	// Es wird geprüft ob der Server noch ausgeführt wird
	if !server_running() {
		return fmt.Errorf("WebsocketKernelServerEP: server starting error")
	}

	// Dere Vorgang wurde ohne Fehler gestartet
	return nil
}

// Wird verwendet um den QUIC basierten Server zu Starten
func (obj *WebsocketKernelServerEP) _start_quic_server() error {
	// Log
	//log.Printf("WebsocketKernelServerEP: new quic based server started. id = %s, endpoint = %s\n", obj._obj_id, "0.0.0.0:"+strconv.Itoa(obj._port))
	return nil
}

// Wird verwendet um zu ermitteln ob einer der Server ausgeführt wird
func (obj *WebsocketKernelServerEP) _wait_of_rsock() {
	// Prüft ob mindestens einen ein Protokoll Thread läuft
	fc := func(tbo *WebsocketKernelServerEP) bool {
		// Der Threadlock wird gesperrt
		tbo._lock.Lock()

		// Es wird geprüft ob mindestens 1 Protokoll Thread vorhanden ist
		total := tbo._total_proc

		// Der Thradlock wird wieder freigegeben
		tbo._lock.Unlock()

		// Der Status wird zurückgegeben
		return total > 0
	}

	// Wird solange ausgeführt bis kein Protokoll mehr ausgeführt wird
	for fc(obj) {
		time.Sleep(1 * time.Millisecond)
	}
}

// Startet den eigentlichen Server
func (obj *WebsocketKernelServerEP) Start() error {
	// Die Channel werden zurückgegeben
	result_chan := make(chan error)

	// Es wird eine neuer Thread gestartet um die Einzelnen Server Protokollthreads zu starten
	go func(wsbo *WebsocketKernelServerEP, state chan error) {
		// Die Server werden gestartet
		tcp_err := obj._start_tcp_tcp_server()
		quic_err := obj._start_quic_server()

		// Es wird Signalisiert dass der Server läuft
		wsbo._lock.Lock()

		// Es wird geprüft ob mindestens 1 Server gestartet wurde
		if tcp_err != nil && quic_err != nil {
			// Der Threadlock wird freigegeben
			wsbo._lock.Unlock()

			// Der Fehler wird zurückgegeben
			state <- fmt.Errorf("neither the TCP nor the QUIC Server Socket could be created")

			// Der Vorgang wird beendet
			return
		}

		// Es wird zurückgegeben dass kein Fehler aufgetreten ist
		state <- nil

		// Es wird Signalisiert dass der Acceptor Thread ausgeführt wird
		wsbo._is_running = true

		// Der Threadlock wird freigegeben
		wsbo._lock.Unlock()

		// Wird solange ausgeführt bis alle Sockets geschlossen wurden
		wsbo._wait_of_rsock()

		// Der Threadlock wird ausgeführt
		wsbo._lock.Lock()

		// Es wird Signalisiert dass der Thread nicht mehr läuft
		wsbo._is_running = false

		// Der Threadlock wird freigegeben
		wsbo._lock.Unlock()
	}(obj, result_chan)

	// Es wird auf die Antwort der Gegenseite gewartet
	res := <-result_chan

	// Sollte ein Fehler aufgetreten sein, wird dieser Zurückgegeben
	if res != nil {
		return res
	}

	// Es wird gewartet bis der Server gestartet wurde
	for !obj._is_rn() {
		time.Sleep(1 * time.Millisecond)
	}

	// Es wird 10 MS gewartet
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Millisecond)
	}

	// Der Threadlock wird gesperrt
	obj._lock.Lock()

	// Es wird geprüft ob mindestens einer der Server läuft
	if !obj._is_running {
		// Der Threadlock wird freiegegeben
		obj._lock.Unlock()

		// Der Fehler wird zurückgegeben
		return fmt.Errorf("WebsocketKernelServerEP: Starting error")
	}

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Der Vorgang wurde ohne Fehler gestartet
	return nil
}

// Gibt das Aktuelle Protokoll aus
func (obj *WebsocketKernelServerEP) GetProtocol() string {
	return "ws"
}

// Gibt die Aktuelle Objekt ID aus
func (obj *WebsocketKernelServerEP) GetObjectId() string {
	return obj._obj_id
}

// Gibt an ob der Server bereits gestartet wurde
func (obj *WebsocketKernelServerEP) IsRunning() bool {
	return false
}

// Upgradet die HTTP Verbindung und erstellt eine Client Sitzung daraus
func (obj *WebsocketKernelServerEP) upgradeHTTPConnAndRegister(w http.ResponseWriter, r *http.Request) {
	// Die Verbindung wird zu einer Websocket Verbindung geupgradet zu einer Websocket verbindung
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		r.Body.Close()
		return
	}

	// Die Aktuelle Zeit wird ermittelt
	c_time := time.Now()

	// Es wird auf das eintreffende Paket gewartet
	mtype, message, err := conn.ReadMessage()
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Es wird geprüft ob es sich um ein Byte Message handelt
	if mtype != websocket.BinaryMessage {
		conn.Close()
		return
	}

	// Es wird versucht das Paket mit dem Lokalen Schlüssel zu entschlüsseln
	decrypted, err := obj._kernel.DecryptWithPrivateRelayKey(message)
	if err != nil {
		fmt.Println("Server", err)
		conn.Close()
		return
	}

	// Es wird versucht den Datensatz wieder einzulesen
	var decrypted_chpackage EncryptedClientHelloPackage
	if err := cbor.Unmarshal(decrypted, &decrypted_chpackage); err != nil {
		fmt.Println("error:", err)
		r.Body.Close()
		return
	}

	// Es wird geprüft ob der Schlüssel des Aktuellen Relays mit dem Angeforderten übereinstiemmt
	if !bytes.Equal(obj._kernel.GetPublicKey().SerializeCompressed(), decrypted_chpackage.PublicServerKey) {
		r.Body.Close()
		return
	}

	// Es wird versucht die Öffentlichen Schlüssel einzulesen
	pub_client_key, err := utils.ReadPublicKeyFromByteSlice(decrypted_chpackage.PublicClientKey)
	if err != nil {
		r.Body.Close()
		return
	}
	pub_client_otk_key, err := utils.ReadPublicKeyFromByteSlice(decrypted_chpackage.RandClientPKey)
	if err != nil {
		r.Body.Close()
		return
	}

	// Der Hash zum überprüfen der Signatur wird erstellt
	sign_hash := utils.ComputeSha3256Hash(decrypted_chpackage.PublicServerKey, decrypted_chpackage.RandClientPKey)

	// Es wird geprüft ob die Signatur korrekt ist
	check, err := utils.VerifyByBytes(pub_client_key, decrypted_chpackage.ClientSig, sign_hash)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !check {
		fmt.Println("Invalid sig")
		return
	}

	// Es wird ein Temporäres Schlüsselpaar erstellt
	key_pair_id, err := obj._kernel.CreateNewTempKeyPair()
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Der Öffentliche Schlüssel wird abgerufen
	temp_public_key, err := obj._kernel.GetPublicTempKeyById(key_pair_id)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Es wird ein Hash zum signieren erstellt 'SHA3_256(decoded_pkey || temp_public_key)'
	serve_sign_hash := utils.ComputeSha3256Hash(decrypted_chpackage.PublicClientKey, temp_public_key.SerializeCompressed(), obj._kernel.GetPublicKey().SerializeCompressed())

	// Der Hash wird mit dem Relay Schlüssel des Aktuellen Relays Signiert
	relay_signature, err := obj._kernel.SignWithRelayKey(serve_sign_hash)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Der Hash wird mit dem Temprären Schlüssel signiert
	temp_key_signature, err := obj._kernel.SignWithTempKeyId(key_pair_id, sign_hash)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Das Antwortpaket wird gebaut
	plain_tcp_server_hello_package := EncryptedServerHelloPackage{
		PublicServerKey:   obj._kernel.GetPublicKey().SerializeCompressed(),
		PublicClientKey:   decrypted_chpackage.PublicClientKey,
		RandServerPKey:    temp_public_key.SerializeCompressed(),
		ServerSig:         relay_signature,
		RandServerPKeySig: temp_key_signature,
	}

	// Das Paket wird in Bytes umgewandelt
	byted, err := cbor.Marshal(plain_tcp_server_hello_package, cbor.EncOptions{})
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Die Daten werden mit dem Öffentlichen Schlüssel der gegenseite verschlüsselt
	encrypted_package, err := utils.EncryptECIESPublicKey(pub_client_key, byted)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Es wird geprüft ob es sich um einen bekannten Relay handelt
	relay_obj, err := obj._kernel.GetTrustedRelayByPublicKey(pub_client_key)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Solte kein Vertrauenswürdiger Relay vorhanden sein, wird ein Temporärer Relay erzeugt
	if relay_obj == nil {
		c_time := time.Now().Unix()
		relay_obj = kernel.NewUntrustedRelay(pub_client_key, c_time, r.Host, "ws")
	}

	// Es wird ein ECDH Schlüssel für die OTK Schlüssel beider Relays erstellt
	otk_ecdh_key, err := obj._kernel.CreateOTKECDHKey(key_pair_id, pub_client_otk_key)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Die Daten werden übermittelt
	send_err := conn.WriteMessage(websocket.BinaryMessage, encrypted_package)
	if send_err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Zeitdifferenz berechnen
	total_ts_time := time.Until(c_time).Seconds()

	// Bandbreite berechnen
	bandwith_kbs := float64(float64(len(message))/total_ts_time) / 1024

	// Das Verbindungsobjekt wird erstellt
	conn_obj, err := createFinallyKernelConnection(conn, key_pair_id, pub_client_key, pub_client_otk_key, otk_ecdh_key, bandwith_kbs, uint64(total_ts_time), kernel.OUTBOUND)
	if err != nil {
		conn.Close()
		log.Println("error: ", err.Error())
		return
	}

	// Die Verbindung wird registriert
	if err := obj._kernel.AddNewConnection(relay_obj, conn_obj); err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Die Verbindung wird final fertigestellt
	if err := conn_obj.FinallyInit(); err != nil {
		obj._kernel.RemoveConnection(conn_obj)
		conn.Close()
	}
}

// Erstellt einen neuen Lokalen Websocket Server
func CreateNewLocalWebsocketServerEP(ip_adr string, port uint64) (*WebsocketKernelServerEP, error) {
	// Die Einmalige ObjektID wird erstellt
	rand_id := utils.RandStringRunes(16)

	// Das Objekt wird vorbereitet
	result_obj := &WebsocketKernelServerEP{_obj_id: rand_id, _lock: new(sync.Mutex), _total_proc: 0, _port: int(port)}

	// Es wird eine zufälliger Objekt ID erstellt
	log.Printf("New Websocket Server EndPoint on %s and port %d created\n", ip_adr, port)
	return result_obj, nil
}
