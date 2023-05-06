package ipoverlay

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fluffelpuff/RoueX/utils"
	"github.com/fxamacker/cbor"
	"github.com/gorilla/websocket"
)

// Stellt eine Websocket Kernel Verbindung dar
type WebsocketKernelClient struct {
	_final_session *WebsocketKernelConnection
	_lock          sync.Mutex
	_kernel        *kernel.Kernel
	_opproc        bool
	_obj_id        string
}

// Stellt Proxy Einstellungen für eine Client Verbindung dar
type WebsocketKernelProxySettings struct {
}

// Setz den Offennen Versuch zurück
func (obj *WebsocketKernelClient) _reset_proc() {
	obj._lock.Lock()
	obj._opproc = false
	obj._lock.Unlock()
}

// Gibt an ob ein Offener Prozess vorhanden ist
func (obj *WebsocketKernelClient) _has_oproc() bool {
	obj._lock.Lock()
	r := obj._opproc
	obj._lock.Unlock()
	return r
}

// Gibt an ob der Client verbunden ist
func (obj *WebsocketKernelClient) _is_connected() bool {
	obj._lock.Lock()
	r := obj._final_session
	obj._lock.Unlock()
	if r == nil {
		return false
	}
	return r.IsConnected()
}

// Registriert den Kernel im Modul
func (obj *WebsocketKernelClient) RegisterKernel(kernel *kernel.Kernel) error {
	obj._kernel = kernel
	return nil
}

// Gibt alle Meta Daten des Moduls aus
func (obj *WebsocketKernelClient) GetMetaDataInfo() kernel.ClientModuleMetaData {
	return kernel.ClientModuleMetaData{}
}

// Gibt das Protokoll des Moduls aus
func (obj *WebsocketKernelClient) GetProtocol() string {
	return "ws"
}

// Stellt eine neue Websocket Verbindung her
func (obj *WebsocketKernelClient) ConnectTo(url_str string, pub_key *btcec.PublicKey, proxy_config *kernel.ProxyConfig) error {
	// Es wird geprüft ob es derzeit eine Verbindung gibt
	if obj._has_oproc() {
		return fmt.Errorf("ConnectTo: always open connecting process")
	}
	if obj._is_connected() {
		return fmt.Errorf("ConnectTo: always connected")
	}

	// Es wird versucht eine Websocket verbindung aufzubauen
	var err error
	var conn *websocket.Conn
	if proxy_config != nil {
		// Set the HTTP proxy to use
		proxyURL, err := url.Parse(proxy_config.Host)
		if err != nil {
			obj._reset_proc()
			return err
		}

		// Create a custom Dialer that uses the HTTP proxy
		dialer := &websocket.Dialer{
			Proxy: http.ProxyURL(proxyURL),
			NetDial: func(network, addr string) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		}

		// Log
		log.Printf("WebsocketKernelClient: trying to establish a websocket connection to trought proxy %s -- %s \n", url_str, hex.EncodeToString(pub_key.SerializeCompressed()))

		// Die Verbindung wird aufgebaut
		conn, _, err = dialer.Dial(url_str, nil)
		if err != nil {
			obj._reset_proc()
			return fmt.Errorf(err.Error())
		}

		// Log
		log.Printf("WebsocketKernelClient: the websocket base connection to %s has been established.\n", url_str)
	} else {
		// Log
		log.Printf("WebsocketKernelClient: trying to establish a websocket connection to %s -- %s\n", url_str, hex.EncodeToString(pub_key.SerializeCompressed()))

		// Die Verbindung wird aufgebaut
		conn, _, err = websocket.DefaultDialer.Dial(url_str, nil)
		if err != nil {
			obj._reset_proc()
			return fmt.Errorf(err.Error())
		}

		// Log
		log.Printf("WebsocketKernelClient: the websocket base connection to %s has been established.\n", url_str)
	}

	// Es wird ein Temporäres Schlüsselpaar erstellt
	key_pair_id, err := obj._kernel.CreateNewTempKeyPair()
	if err != nil {
		obj._reset_proc()
		return fmt.Errorf("ConnectTo: 5: " + err.Error())
	}

	// Der Öffentliche Schlüssel wird abgerufen
	temp_public_key, err := obj._kernel.GetPublicTempKeyById(key_pair_id)
	if err != nil {
		obj._reset_proc()
		return fmt.Errorf("ConnectTo: 2: " + err.Error())
	}

	// Es wird geprüft ob es sich um einen bekannten Relay handelt
	relay_pkyobj, err := obj._kernel.GetTrustedRelayByPublicKey(pub_key)
	if err != nil {
		obj._reset_proc()
		log.Println(err)
		conn.Close()
		return nil
	}

	// Es wird ein Hash zum signieren erstellt 'SHA3_256(decoded_pkey || temp_public_key)'
	sign_hash := kernel.ComputeSha3256Hash(pub_key.SerializeCompressed(), temp_public_key.SerializeCompressed())

	// Der Hash wird mit dem Relay Schlüssel des Aktuellen Relays Signiert
	relay_signature, err := obj._kernel.SignWithRelayKey(sign_hash)
	if err != nil {
		obj._reset_proc()
		return fmt.Errorf("ConnectTo: 3: " + err.Error())
	}

	// Der Hash wird mit dem Temprären Schlüssel signiert
	temp_key_signature, err := obj._kernel.SignWithTempKeyId(key_pair_id, sign_hash)
	if err != nil {
		obj._reset_proc()
		return fmt.Errorf("ConnectTo: 4: " + err.Error())
	}

	// Die Aktuelle Zeit wird ermittelt
	c_time := time.Now()

	// Das Verschlüsselt HelloClientPackage wird vorbereitet
	plain_client_hello_package := EncryptedClientHelloPackage{
		PublicServerKey:   pub_key.SerializeCompressed(),
		PublicClientKey:   obj._kernel.GetPublicKey().SerializeCompressed(),
		RandClientPKey:    temp_public_key.SerializeCompressed(),
		ClientSig:         relay_signature,
		RandClientPKeySig: temp_key_signature,
	}

	// Das Paket wird in Bytes umgewandelt
	byted, err := cbor.Marshal(plain_client_hello_package, cbor.EncOptions{})
	if err != nil {
		obj._reset_proc()
		return err
	}

	// Die Daten werden mit dem Öffentlichen Schlüssel der gegenseite verschlüsselt
	encrypted_package, err := kernel.EncryptECIESPublicKey(pub_key, byted)
	if err != nil {
		obj._reset_proc()
		return fmt.Errorf("ConnectTo: 6: " + err.Error())
	}

	// Die Daten werden übermittelt
	send_err := conn.WriteMessage(websocket.BinaryMessage, encrypted_package)
	if send_err != nil {
		obj._reset_proc()
		return err
	}

	// Es wird Maximal 120 Sekunden auf die Antwort gewartet
	timeout := 120 * time.Second
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Es wird auf die Antwort gewartet
	messageType, recived_message, err := conn.ReadMessage()
	if err != nil {
		obj._reset_proc()
		return err
	}

	// Sollte es sich nicht um eine Binäry Message handelt, wird der Vorgang abgebrochen und die Verbindung wird geschlossen
	if messageType != websocket.BinaryMessage {
		obj._reset_proc()
		return fmt.Errorf("internal error, unkown response from another relay")
	}

	// Es wird versucht den Datensatz mit dem Private Relay Schlüssel zu entschlüsseln
	decrypted_message, err := obj._kernel.DecryptWithPrivateRelayKey(recived_message)
	if err != nil {
		obj._reset_proc()
		return err
	}

	// Es wird versucht die Daten mittels CBOR einzulesen
	var eshp EncryptedServerHelloPackage
	if err := cbor.Unmarshal(decrypted_message, &eshp); err != nil {
		obj._reset_proc()
		return err
	}

	// Es wird versucht den Öffentlicher Schlüssel des Servers einzulesen
	public_server_key, err := kernel.ReadPublicKeyFromByteSlice(eshp.PublicServerKey)
	if err != nil {
		obj._reset_proc()
		return err
	}
	public_server_otk, err := kernel.ReadPublicKeyFromByteSlice(eshp.RandServerPKey)
	if err != nil {
		obj._reset_proc()
		return err
	}

	// Das Reading Timeout wird entfernt
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		obj._reset_proc()
		return err
	}

	// Es wird ein ECDH Schlüssel für die OTK Schlüssel beider Relays erstellt
	otk_ecdh_key, err := obj._kernel.CreateOTKECDHKey(key_pair_id, public_server_otk)
	if err != nil {
		obj._reset_proc()
		return err
	}

	// Zeitdifferenz berechnen
	total_ts_time := time.Until(c_time).Seconds()

	// Bandbreite berechnen
	bandwith_kbs := float64(float64(len(recived_message))/total_ts_time) / 1024

	// Das Finale Sitzungsobjekt wird erstellt
	finally_kernel_session, err := createFinallyKernelConnection(conn, key_pair_id, public_server_key, public_server_otk, otk_ecdh_key, bandwith_kbs, uint64(total_ts_time), kernel.OUTBOUND)
	if err != nil {
		obj._reset_proc()
		conn.Close()
		return err
	}

	// Solte kein Vertrauenswürdiger Relay vorhanden sein, wird ein Temporärer Relay erzeugt
	if relay_pkyobj == nil {
		c_time := time.Now().Unix()
		relay_pkyobj = kernel.NewUntrustedRelay(pub_key, c_time, url_str, "ws")
		log.Println("Unkown relay", relay_pkyobj.GetHexId(), "connected")
	} else {
		log.Println("Trusted relay", relay_pkyobj.GetHexId(), "connected")
	}

	// Die Verbindung wird registriert
	if err := obj._kernel.AddNewConnection(relay_pkyobj, finally_kernel_session); err != nil {
		obj._reset_proc()
		conn.Close()
		return err
	}

	// Die Verbindung wird final fertigestellt
	if err := finally_kernel_session.FinallyInit(); err != nil {
		obj._kernel.RemoveConnection(finally_kernel_session)
		obj._reset_proc()
		conn.Close()
		return err
	}

	// Das Finale Objekt wird abgespeichert
	obj._lock.Lock()
	obj._opproc = false
	obj._final_session = finally_kernel_session
	obj._lock.Unlock()

	// Der Gegenseite wird nun der eigene Öffentliche Schlüssel, die Aktuelle Uhrzeit sowie
	return nil
}

// Gibt die Aktuelle ObjektID aus
func (obj *WebsocketKernelClient) GetObjectId() string {
	return obj._obj_id
}

// Beendet das Module, verhindert das weitere verwenden
func (obj *WebsocketKernelClient) Shutdown() {
	log.Println("Shutdowing websocket clients")
}

// Wird verwendet wenn es sich um eine Ausgehende Verbindung handelt
func (obj *WebsocketKernelClient) Serve() {
	obj._lock.Lock()
	if obj._final_session == nil {
		obj._lock.Unlock()
		return
	}
	obj._lock.Unlock()
	for obj._final_session.IsConnected() {
		time.Sleep(1 * time.Millisecond)
	}
}

// Erstellt ein neues Websocket Client Modul
func NewWebsocketClient() *WebsocketKernelClient {
	rand_id := utils.RandStringRunes(16)
	return &WebsocketKernelClient{_obj_id: rand_id, _lock: sync.Mutex{}}
}
