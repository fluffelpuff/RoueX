package ipoverlay

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fxamacker/cbor"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type WebsocketKernelServerEP struct {
	_kernel *kernel.Kernel
	_obj_id string
}

// Registriert den Kernel im Module
func (obj *WebsocketKernelServerEP) RegisterKernel(k *kernel.Kernel) error {
	log.Printf("Websocket Server EndPoint registrated on kernel %s\n", k.GetKernelID())
	obj._kernel = k
	return nil
}

// Wird verwendet um den Serversocket herunterzufahren
func (obj *WebsocketKernelServerEP) Shutdown() {
	log.Printf("Websocket Server EndPoint shutingdown...\n")
}

// Startet den eigentlichen Server
func (obj *WebsocketKernelServerEP) Start() error {
	log.Printf("New Websocket Server EndPoint started\n")
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
	relay_pkyobj, err := obj._kernel.GetTrustedRelayByPublicKey(pub_client_key)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Solte kein Vertrauenswürdiger Relay vorhanden sein, wird ein Temporärer Relay erzeugt
	if relay_pkyobj == nil {
		c_time := time.Now().Unix()
		relay_pkyobj = kernel.NewUntrustedRelay(pub_client_key, c_time, r.Host, "ws")
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
	conn_obj, err := createFinallyKernelConnection(conn, key_pair_id, pub_client_key, pub_client_otk_key, bandwith_kbs, int64(total_ts_time))
	if err != nil {
		conn.Close()
		log.Println("error: ", err.Error())
		return
	}

	// Die Verbindung wird registriert
	if err := obj._kernel.AddNewConnection(relay_pkyobj, conn_obj); err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	// Die Verbindung wird final fertigestellt
	if err := conn_obj.FinallyInit(); err != nil {
		obj._kernel.RemoveConnection(relay_pkyobj, conn_obj)
		conn.Close()
	}

	// Die bekannten Routen für diese Verbindung (Relay) werden abgerufen
	if err := obj._kernel.LoadAndActiveRoutesByRelay(relay_pkyobj); err != nil {
		obj._kernel.RemoveConnection(relay_pkyobj, conn_obj)
		log.Println("error:", err.Error())
		conn.Close()
		return
	}
}

// Erstellt einen neuen Lokalen Websocket Server
func CreateNewLocalWebsocketServerEP(ip_adr string, port uint64) (*WebsocketKernelServerEP, error) {
	// Die Einmalige ObjektID wird erstellt
	rand_id := kernel.RandStringRunes(16)

	// Das Objekt wird vorbereitet
	result_obj := &WebsocketKernelServerEP{_obj_id: rand_id}

	// Es wird eine neuer Thread gestartet, innerhalb dieses Threads wird der HTTP Server ausgeführt
	go func() {
		server := &http.Server{
			Addr:    ":" + strconv.Itoa(int(port)),
			Handler: http.HandlerFunc(result_obj.upgradeHTTPConnAndRegister),
		}

		// Der Server wird ausgeführt
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Es wird eine zufälliger Objekt ID erstellt
	log.Printf("New Websocket Server EndPoint on %s and port %d created\n", ip_adr, port)
	return result_obj, nil
}
