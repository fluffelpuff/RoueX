package ipoverlay

import (
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fxamacker/cbor"
	"github.com/gorilla/websocket"
)

type WebsocketKernelClient struct {
	_kernel *kernel.Kernel
	_obj_id string
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
func (obj *WebsocketKernelClient) ConnectTo(url string, pub_key *btcec.PublicKey) error {
	// Log
	fmt.Printf("Trying to establish a websocket connection to %s -- %s\n", url, pub_key)

	// Es wird versucht eine Websocket verbindung aufzubauen
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("ConnectTo: 1: " + err.Error())
	}

	// Log
	fmt.Printf("The websocket base connection to %s has been established.\n", url)

	// Es wird ein Temporäres Schlüsselpaar erstellt
	key_pair_id, err := obj._kernel.CreateNewTempKeyPair()
	if err != nil {
		return fmt.Errorf("ConnectTo: 5: " + err.Error())
	}

	// Der Öffentliche Schlüssel wird abgerufen
	temp_public_key, err := obj._kernel.GetPublicTempKeyById(key_pair_id)
	if err != nil {
		return fmt.Errorf("ConnectTo: 2: " + err.Error())
	}

	// Es wird ein Hash zum signieren erstellt 'SHA3_256(decoded_pkey || temp_public_key)'
	sign_hash := kernel.ComputeSha3256Hash(pub_key.SerializeCompressed(), temp_public_key.SerializeCompressed())

	// Der Hash wird mit dem Relay Schlüssel des Aktuellen Relays Signiert
	relay_signature, err := obj._kernel.SignWithRelayKey(sign_hash)
	if err != nil {
		return fmt.Errorf("ConnectTo: 3: " + err.Error())
	}

	// Der Hash wird mit dem Temprären Schlüssel signiert
	temp_key_signature, err := obj._kernel.SignWithTempKeyId(key_pair_id, sign_hash)
	if err != nil {
		return fmt.Errorf("ConnectTo: 4: " + err.Error())
	}

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
		return err
	}

	// Die Daten werden mit dem Öffentlichen Schlüssel der gegenseite verschlüsselt
	encrypted_package, err := kernel.EncryptECIESPublicKey(pub_key, byted)
	if err != nil {
		return fmt.Errorf("ConnectTo: 6: " + err.Error())
	}

	// Die Zeit zum zeitpunkt des Absendens wird ermittelt
	send_timestamp := time.Now()

	// Die Daten werden übermittelt
	send_err := conn.WriteMessage(websocket.BinaryMessage, encrypted_package)
	if send_err != nil {
		return err
	}

	// Es wird Maximal 120 Sekunden auf die Antwort gewartet
	timeout := 120 * time.Second
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Es wird auf die Antwort gewartet
	messageType, recived_message, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	// Die Aktuelle Zeit wird ermittelt
	recived_timestamp := time.Now()

	// Sollte es sich nicht um eine Binäry Message handelt, wird der Vorgang abgebrochen und die Verbindung wird geschlossen
	if messageType != websocket.BinaryMessage {
		return fmt.Errorf("internal error, unkown response from another relay")
	}

	// Es wird versucht den Datensatz mit dem Private Relay Schlüssel zu entschlüsseln
	decrypted_message, err := obj._kernel.DecryptWithPrivateRelayKey(recived_message)
	if err != nil {
		return err
	}

	// Es wird versucht die Daten mittels CBOR einzulesen
	var eshp EncryptedServerHelloPackage
	if err := cbor.Unmarshal(decrypted_message, &eshp); err != nil {
		return err
	}

	// Es wird versucht den Öffentlicher Schlüssel des Servers einzulesen
	public_server_key, err := kernel.ReadPublicKeyFromByteSlice(eshp.PublicServerKey)
	if err != nil {
		return err
	}
	public_server_otk, err := kernel.ReadPublicKeyFromByteSlice(eshp.RandServerPKey)
	if err != nil {
		return err
	}

	// Das Reading Timeout wird entfernt
	if err := conn.SetReadDeadline(time.Unix(0, 0)); err != nil {
		return err
	}

	// Das Finale Sitzungsobjekt wird erstellt
	finally_kernel_session, err := createFinallyKernelConnection(conn, key_pair_id, public_server_key, public_server_otk)
	if err != nil {
		conn.Close()
		return err
	}

	// Berechnen Sie die Ping-Zeit als Zeitunterschied zwischen dem Senden und Empfangen der Nachricht
	pingTime := recived_timestamp.Sub(send_timestamp)

	// Die Banbreite wird ausgerechnet
	bandwidth := fmt.Sprintf("%.2f", float64(len(encrypted_package)+len(recived_message))/(pingTime.Seconds()*1024))

	fmt.Println(finally_kernel_session, bandwidth)

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

// Erstellt ein neues Websocket Client Modul
func NewWebsocketClient() *WebsocketKernelClient {
	rand_id := kernel.RandStringRunes(16)
	return &WebsocketKernelClient{_obj_id: rand_id}
}
