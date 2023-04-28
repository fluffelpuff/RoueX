package kernel

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"plugin"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/fluffelpuff/RoueX/static"
)

// Stellt das Kernel Objekt dar
type Kernel struct {
	_external_modules_path string
	_os_path_trimmer       string
	_socket_path           string
	_kernel_id             string
	_is_running            bool
	_lock                  *sync.Mutex
	_socket                net.Listener
	_routing_table         *RoutingTable
	_trusted_relays        *TrustedRelays
	_server_modules        []*ServerModule
	_client_modules        []*ClientModule
	_connection_manager    ConnectionManager
	_firewall              *Firewall
	_private_key           *btcec.PrivateKey
	_temp_key_pairs        map[string]*btcec.PrivateKey
}

// Erstellt einen OSX Kernel
func CreateOSXKernel(priv_key string) (*Kernel, error) {
	// Log
	fmt.Println("Creating new RoueX OSX Kernel...")

	// Der Öffentliche Schlüssel wird dekodiert
	decoded_priv_key, err := hex.DecodeString(priv_key)
	if err != nil {
		return nil, err
	}

	// Sollte der Private Schlüssel keine 32 Bytes groß sein wird der Vorgang abgebrochen
	if len(decoded_priv_key) != 32 {
		return nil, fmt.Errorf("invalid private key length")
	}

	// Der Private Schlüssel wird als Objekt eingelesen
	dec_object, err := ReadPrivateKeyFromByteSlice(decoded_priv_key)
	if err != nil {
		return nil, err
	}

	// Es wird eine Liste mit allen Vertrauten Relays abgerufen
	trusted_relays_obj, err := loadTrustedRelaysTable(static.OSX_TRUSTED_RELAYS_PATH)
	if err != nil {
		log.Fatal("listen error:", err)
		return nil, err
	}

	// Es wird versucht die Routing Tabelle zu laden
	routing_table_obj, err := loadRoutingTable(static.OSX_ROUTING_TABLE_PATH)
	if err != nil {
		log.Fatal("listen error:", err)
		return nil, err
	}

	// Es wird versucht die Firewall Tabelle zu ladne
	firewall_table_obj, err := loadFirewallTable(static.OSX_FIREWALL_TABLE_PATH)
	if err != nil {
		log.Fatal("listen error:", err)
		return nil, err
	}

	// Der Unix Socket wird vorbereitet
	l, err := net.Listen("unix", static.OSX_NO_ROOT_API_SOCKET)
	if err != nil {
		log.Fatal("listen error:", err)
		return nil, err
	}

	// Log
	fmt.Println("RoueX OSX Kernel API Unix-Socket created...")

	// Die KernelID wird estellt
	k_id := RandStringRunes(16)

	// Die Verbindungsverwaltung wird erstellt
	conn_manager := newConnectionManager()

	// Erstellt das Kernel Objekt
	new_kernel := Kernel{
		_socket:                l,
		_os_path_trimmer:       "/",
		_kernel_id:             k_id,
		_connection_manager:    conn_manager,
		_private_key:           dec_object,
		_lock:                  new(sync.Mutex),
		_temp_key_pairs:        make(map[string]*secp256k1.PrivateKey),
		_socket_path:           static.OSX_NO_ROOT_API_SOCKET,
		_routing_table:         &routing_table_obj,
		_firewall:              firewall_table_obj,
		_trusted_relays:        &trusted_relays_obj,
		_external_modules_path: static.OSX_EXTERNAL_MODULES,
	}

	// Gibt das Kernelobjekt ohne Fehler zurück
	fmt.Printf("New OSX RoueX Kernel '%s' created...\n", k_id)
	return &new_kernel, nil
}

// Wird unter OSX, Linux oder Windows verwendet zum aufräumen
func (obj *Kernel) CleanUp() error {
	log.Println("Clearing kernel...")
	return nil
}

// Wird ausgeführt wenn das Programm als Dienst ausgeführt wird
func (obj *Kernel) Start() error {
	// Der Lokale nicht Root CLI Socket wird erstellt
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			os.Remove(obj._socket_path)
		}()

		obj._lock.Lock()
		obj._is_running = true
		obj._lock.Unlock()
		errChan <- nil

		log.Println("Kernel started...")
		rpc.HandleHTTP()
		rpc.Accept(obj._socket)

		obj._lock.Lock()
		obj._is_running = false
		obj._lock.Unlock()
	}()
	r := <-errChan
	if r != nil {
		return r
	}
	return nil
}

// Wird ausgeführt um den Kernel zu beenden
func (obj *Kernel) Shutdown() {
	var is_shutdown bool
	obj._lock.Lock()
	is_shutdown = obj._is_running
	obj._lock.Unlock()

	if is_shutdown {
		// Die Internen Dienste werden beendet
		var vat ServerModule
		for _, item := range obj._server_modules {
			if item == nil {
				continue
			}
			vat = *item
			vat.Shutdown()
		}

		// Die Datenbanken werden geschlossen
		obj._routing_table.Shutdown()
		obj._trusted_relays.Shutdown()

		// Der RPC Unix Socket wird geschlossen
		obj._socket.Close()

		// Log
		log.Println("Kernel shutdown...")
	}
}

// Gibt an ob der Kernel ausgeführt wird
func (obj *Kernel) IsRunning() bool {
	var is_running bool
	obj._lock.Lock()
	is_running = obj._is_running
	obj._lock.Unlock()
	return is_running
}

// Fügt einen Lokalen Server Endpunkt hinzu
func (obj *Kernel) RegisterServerModule(lcsep ServerModule) error {
	if obj.IsRunning() {
		return fmt.Errorf("can't add local ep than server is running")
	}
	log.Printf("Register new server module, protocol = %s, id = %s\n", lcsep.GetProtocol(), lcsep.GetObjectId())
	if err := lcsep.RegisterKernel(obj); err != nil {
		return err
	}
	obj._server_modules = append(obj._server_modules, &lcsep)
	return nil
}

// Fügt eine Clientfunktion hinzu, diese Erlaubt ausgehende Verbindungen
func (obj *Kernel) RegisterClientModule(csep ClientModule) error {
	if obj.IsRunning() {
		return fmt.Errorf("can't add local ep than server is running")
	}
	log.Printf("Register new client module, protocol = %s, id = %s\n", csep.GetProtocol(), csep.GetObjectId())
	if err := csep.RegisterKernel(obj); err != nil {
		return err
	}
	obj._client_modules = append(obj._client_modules, &csep)
	return nil
}

// Gibt eine Liste mit allen Verfügbaren Relays zurück
func (obj *Kernel) GetTrustedRelays() ([]*Relay, error) {
	return nil, nil
}

// Gibt einen Relay anhand seinens PublicKeys zurück
func (obj *Kernel) GetTrustedRelayByPublicKey(pkey *btcec.PublicKey) (*Relay, error) {
	_, err := obj.GetTrustedRelays()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Markiert einen Relay als Verbunden
func (obj *Kernel) AddNewConnection(relay *Relay, conn RelayConnection) error {
	// Sollte kein Relay vorhanden sein, wird die Verbindung als nicht Verifiziert gespeichert
	if err := obj._connection_manager.RegisterNewRelayConnection(relay, conn); err != nil {
		return err
	}

	// Der Vorgang wurde erfolgreich druchgeführt
	return nil
}

// Markiert einen Relay als nicht mehr Verbunden
func (obj *Kernel) RemoveConnection(relay *Relay, conn RelayConnection) error {
	// Sollte kein Relay vorhanden sein, wird die Verbindung als nicht Verifiziert gespeichert
	if err := obj._connection_manager.RemoveRelayConnection(relay, conn); err != nil {
		return err
	}

	// Der Vorgang wurde erfolgreich druchgeführt
	return nil
}

// Wird verwendet um Third Party oder Externe Kernel Module zu laden
func (obj *Kernel) LoadExternalKernelModules() error {
	files, err := os.ReadDir(obj._external_modules_path)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Loading external Kernel Modules from %s\n", obj._external_modules_path)

	loaded_modules := []*ExternalModule{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		plug, err := plugin.Open(obj._external_modules_path + obj._os_path_trimmer + file.Name())
		if err != nil {
			continue
		}

		lamda_kernel_mod, err := plug.Lookup("Module")
		if err != nil {
			fmt.Println(file.Name(), err)
			continue
		}

		loaded_kernel_module, ok := lamda_kernel_mod.(ExternalModule)
		if !ok {
			fmt.Println("NOT_OK")
			continue
		}

		log.Printf("Kernel Module %s loaded\n", obj._external_modules_path+obj._os_path_trimmer+file.Name())

		if err := loaded_kernel_module.Info(); err != nil {
			fmt.Println(err)
			return err
		}

		loaded_modules = append(loaded_modules, &loaded_kernel_module)
	}

	log.Printf("%d Kernel Modules loaded\n", len(loaded_modules))
	return nil
}

// Wird verwendet um die Externen Kernel Module zu starten
func (obj *Kernel) StartExternalKernelModules() error {
	log.Println("Starting external kernel modules")
	return nil
}

// Gibt die KernelID aus
func (obj *Kernel) GetKernelID() string {
	return obj._kernel_id
}

// Gibt eine Liste mit Verfügabren Relays aus mit denen eine Verbindung möglich ist
func (obj *Kernel) ListOutboundTrustedAvaileRelays() ([]RelayOutboundPair, error) {
	// Es werden alle Endpunkte abgerufen für welches das Protokoll bekannt ist
	filtered_list := []RelayOutboundPair{}
	for _, x := range obj._trusted_relays.GetAllRelays() {
		if len(x._type) > 1 {
			if obj._connection_manager.RelayIsConnected(x) {
				continue
			}
			recov_entry := RelayOutboundPair{_relay: x}
			for _, r := range obj._client_modules {
				vat := *r
				if vat.GetProtocol() == x._type {
					recov_entry._cl_module = r
					break
				}
			}
			if recov_entry._cl_module == nil {
				continue
			}
			filtered_list = append(filtered_list, recov_entry)
		}
	}

	// Es wird eine List mit Vertrauenswürdigen Relays zurückgegeben, mit denen im moment noch keinen Verbindung besteht
	return filtered_list, nil
}

// Gibt den Öffentlichen Schlüssel des Relays aus
func (obj *Kernel) GetPublicKey() *btcec.PublicKey {
	return obj._private_key.PubKey()
}

// Erstellt ein neues Schlüsselpaar, der Zugriff auf den Privaten Schlüssel ist nicht möglich
func (obj *Kernel) CreateNewTempKeyPair() (string, error) {
	rid := randProcId()
	priv_k, err := GeneratePrivateKey()
	if err != nil {
		return "", err
	}

	obj._lock.Lock()
	obj._temp_key_pairs[rid] = priv_k
	obj._lock.Unlock()

	return rid, nil
}

// Gibt einen Öffentlichen Temporären Schlüssel anhand seiner ID aus
func (obj *Kernel) GetPublicTempKeyById(temp_key_id string) (*btcec.PublicKey, error) {
	obj._lock.Lock()
	priv_key, ok := obj._temp_key_pairs[temp_key_id]
	if !ok {
		obj._lock.Unlock()
		return nil, fmt.Errorf("not found")
	}
	obj._lock.Unlock()
	return priv_key.PubKey(), nil
}

// Wird verwendet um eine Signatur mit dem Relay Key zu Signieren
func (obj *Kernel) SignWithRelayKey(digest []byte) ([]byte, error) {
	return Sign(obj._private_key, digest)
}

// Wird verwendet um einen Hash mit Temprären Schlüssel zu Signieren
func (obj *Kernel) SignWithTempKeyId(temp_key_id string, digest []byte) ([]byte, error) {
	obj._lock.Lock()
	priv_key, ok := obj._temp_key_pairs[temp_key_id]
	if !ok {
		obj._lock.Unlock()
		return nil, fmt.Errorf("not found")
	}
	obj._lock.Unlock()
	return Sign(priv_key, digest)
}

// Wird verwendet um einen Verschlüsselten Datensatz mit dem Privaten Relay Schlüssel zu enschlüsseln
func (obj *Kernel) DecryptWithPrivateRelayKey(cipher_data []byte) ([]byte, error) {
	return DecryptDataWithPrivateKey(obj._private_key, cipher_data)
}
