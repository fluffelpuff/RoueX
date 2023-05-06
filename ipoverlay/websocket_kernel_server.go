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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type WebsocketKernelServerEP struct {
	_kernel          *kernel.Kernel
	_obj_id          string
	_shutdown_signal bool
	_is_running      bool
	_server          *http.Server
	_lock            *sync.Mutex
	_port            int
}

// Gibt an ob der Server ausgeführt wird
func (obj *WebsocketKernelServerEP) _is_rn() bool {
	obj._lock.Lock()
	r := obj._is_running
	obj._lock.Unlock()
	return r
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

	// Es wird Signalisiert dass der Server beendet werden soll
	obj._lock.Lock()
	obj._shutdown_signal = true
	obj._server.Close()
	obj._lock.Unlock()

	// Es wird gewartet bis der Server beendet wurde
	for obj._is_rn() {
		time.Sleep(1 * time.Millisecond)
	}
}

// Startet den eigentlichen Server
func (obj *WebsocketKernelServerEP) Start() error {
	// Es wird eine neuer Thread gestartet, innerhalb dieses Threads wird der HTTP Server ausgeführt
	go func(wsbo *WebsocketKernelServerEP) {
		wsbo._server = &http.Server{
			Addr:    ":" + strconv.Itoa(wsbo._port),
			Handler: http.HandlerFunc(wsbo.upgradeHTTPConnAndRegister),
		}

		// Es wird Signalisiert dass der Server läuft
		wsbo._lock.Lock()
		wsbo._is_running = true
		wsbo._lock.Unlock()

		// Der Server wird ausgeführt
		log.Println("WebsocketKernelServerEP: new server started. id =", obj._obj_id)
		if err := wsbo._server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}

		// Es wird Signalisiert dass der Server nicht mehr ausgeführt wird
		wsbo._lock.Lock()
		wsbo._is_running = false
		wsbo._lock.Unlock()

		// Log
		log.Println("WebsocketKernelServerEP: closed. id =", obj._obj_id)
	}(obj)

	// Es wird gewartet bis der Server gestartet wurde
	for !obj._is_rn() {
		time.Sleep(1 * time.Millisecond)
	}

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
	pub_client_key, err := kernel.ReadPublicKeyFromByteSlice(decrypted_chpackage.PublicClientKey)
	if err != nil {
		r.Body.Close()
		return
	}
	pub_client_otk_key, err := kernel.ReadPublicKeyFromByteSlice(decrypted_chpackage.RandClientPKey)
	if err != nil {
		r.Body.Close()
		return
	}

	// Der Hash zum überprüfen der Signatur wird erstellt
	sign_hash := kernel.ComputeSha3256Hash(decrypted_chpackage.PublicServerKey, decrypted_chpackage.RandClientPKey)

	// Es wird geprüft ob die Signatur korrekt ist
	check, err := kernel.VerifyByBytes(pub_client_key, decrypted_chpackage.ClientSig, sign_hash)
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
	serve_sign_hash := kernel.ComputeSha3256Hash(decrypted_chpackage.PublicClientKey, temp_public_key.SerializeCompressed(), obj._kernel.GetPublicKey().SerializeCompressed())

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
	plain_server_hello_package := EncryptedServerHelloPackage{
		PublicServerKey:   obj._kernel.GetPublicKey().SerializeCompressed(),
		PublicClientKey:   decrypted_chpackage.PublicClientKey,
		RandServerPKey:    temp_public_key.SerializeCompressed(),
		ServerSig:         relay_signature,
		RandServerPKeySig: temp_key_signature,
	}

	// Das Paket wird in Bytes umgewandelt
	byted, err := cbor.Marshal(plain_server_hello_package, cbor.EncOptions{})
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Die Daten werden mit dem Öffentlichen Schlüssel der gegenseite verschlüsselt
	encrypted_package, err := kernel.EncryptECIESPublicKey(pub_client_key, byted)
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
	result_obj := &WebsocketKernelServerEP{_obj_id: rand_id, _lock: new(sync.Mutex)}

	// Es wird eine zufälliger Objekt ID erstellt
	log.Printf("New Websocket Server EndPoint on %s and port %d created\n", ip_adr, port)
	return result_obj, nil
}
