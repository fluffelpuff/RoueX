package kernel

import (
	"bytes"
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
	_temp_ecdh_keys        map[string][]byte
}

// Erstellt einen OSX Kernel
func CreateOSXKernel(priv_key *btcec.PrivateKey) (*Kernel, error) {
	// Log
	fmt.Println("Creating new RoueX OSX Kernel...")

	// Es wird eine Liste mit allen Vertrauten Relays abgerufen
	trusted_relays_obj, err := loadTrustedRelaysTable(static.GetFilePathFor(static.TRUSTED_RELAYS))
	if err != nil {
		log.Fatal("listen error:", err)
		return nil, err
	}

	// Es wird versucht die Routing Tabelle zu laden
	routing_table_obj, err := loadRoutingTable(static.GetFilePathFor(static.ROUTING_TABLE))
	if err != nil {
		log.Fatal("listen error:", err)
		return nil, err
	}

	// Es wird versucht die Firewall Tabelle zu ladne
	firewall_table_obj, err := loadFirewallTable(static.GetFilePathFor(static.FIREWALL_TABLE))
	if err != nil {
		log.Fatal("listen error:", err)
		return nil, err
	}

	// Der Unix Socket wird vorbereitet
	l, err := net.Listen("unix", static.GetFilePathFor(static.API_SOCKET))
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
		_private_key:           priv_key,
		_lock:                  new(sync.Mutex),
		_temp_key_pairs:        make(map[string]*secp256k1.PrivateKey),
		_temp_ecdh_keys:        make(map[string][]byte),
		_socket_path:           static.GetFilePathFor(static.API_SOCKET),
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
	return obj._trusted_relays.GetAllRelays(), nil
}

// Gibt einen Relay anhand seinens PublicKeys zurück
func (obj *Kernel) GetTrustedRelayByPublicKey(pkey *btcec.PublicKey) (*Relay, error) {
	relays, err := obj.GetTrustedRelays()
	if err != nil {
		return nil, err
	}
	for i := range relays {
		if bytes.Equal(relays[i]._public_key.SerializeCompressed(), pkey.SerializeCompressed()) {
			return relays[i], nil
		}
	}
	return nil, nil
}

// Markiert einen Relay als Verbunden
func (obj *Kernel) AddNewConnection(relay *Relay, conn RelayConnection) error {
	// Sollte kein Relay vorhanden sein, wird die Verbindung als nicht Verifiziert gespeichert
	if err := obj._connection_manager.RegisterNewRelayConnection(relay, conn); err != nil {
		return err
	}

	// Der Kernel wird in der Verbindung registriert
	if err := conn.RegisterKernel(obj); err != nil {
		// Die Verbindung wird wieder aus dem Verbindungsmanager entfernt
		obj._connection_manager.RemoveRelayConnection(conn)

		// Der Vorgang wird mit einem Fehler abgebrochen
		return err
	}

	// Der Vorgang wurde erfolgreich druchgeführt
	return nil
}

// Markiert einen Relay als nicht mehr Verbunden
func (obj *Kernel) RemoveConnection(relay *Relay, conn RelayConnection) error {
	// Sollte kein Relay vorhanden sein, wird die Verbindung als nicht Verifiziert gespeichert
	if err := obj._connection_manager.RemoveRelayConnection(conn); err != nil {
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
		return nil, fmt.Errorf(fmt.Sprint("not found a", temp_key_id))
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
		return nil, fmt.Errorf(fmt.Sprint("not found b", temp_key_id))
	}
	obj._lock.Unlock()
	return Sign(priv_key, digest)
}

// Erstellt einen OTK ECDH Schlüssel aus einem Öffentlichen Schlüssel und der OTK-ID
func (obj *Kernel) CreateOTKECDHKey(otk_id string, dest_pkey *btcec.PublicKey) (string, error) {
	obj._lock.Lock()
	priv_key, ok := obj._temp_key_pairs[otk_id]
	if !ok {
		obj._lock.Unlock()
		return "", fmt.Errorf(fmt.Sprint("not found c", otk_id))
	}

	shared_secret := btcec.GenerateSharedSecret(priv_key, dest_pkey)

	found_id := ""
	for key := range obj._temp_ecdh_keys {
		if bytes.Equal(obj._temp_ecdh_keys[key], shared_secret) {
			found_id = key
			break
		}
	}
	if len(found_id) > 0 {
		obj._lock.Unlock()
		return found_id, nil
	}

	rand_id := RandStringRunes(12)
	obj._temp_ecdh_keys[rand_id] = shared_secret
	obj._lock.Unlock()

	log.Println("Kernel: new ecdh key computed. otk_id =", rand_id, " dh_hash =", hex.EncodeToString(ComputeSha3256Hash(shared_secret)))
	return rand_id, nil
}

// Verschlüsselt einen Datensatz mit dem OTK ECDH Schlüssel
func (obj *Kernel) EncryptOTKECDHById(algo EncryptionAlgo, otk_id string, data []byte) ([]byte, error) {
	obj._lock.Lock()
	ecdh_key, ok := obj._temp_ecdh_keys[otk_id]
	if !ok {
		obj._lock.Unlock()
		return nil, fmt.Errorf(fmt.Sprint("not found d", otk_id))
	}
	obj._lock.Unlock()

	switch algo {
	case CHACHA_2020:
		r, e := EncryptWithChaCha(ecdh_key, data)
		log.Println("Kernel: encrypting data with chacha20. otk_id =", otk_id, "data_size =", len(data))
		return r, e
	default:
		return nil, fmt.Errorf("unkown algo")
	}
}

// Entschlüsselt einen Datensatz mit dem OTK ECDH Schlüssel
func (obj *Kernel) DecryptOTKECDHById(algo EncryptionAlgo, otk_id string, data []byte) ([]byte, error) {
	obj._lock.Lock()
	ecdh_key, ok := obj._temp_ecdh_keys[otk_id]
	if !ok {
		obj._lock.Unlock()
		return nil, fmt.Errorf(fmt.Sprint("not found d", otk_id))
	}
	obj._lock.Unlock()

	switch algo {
	case CHACHA_2020:
		r, e := DecryptWithChaCha(ecdh_key, data)
		log.Println("Kernel: decrypting data with chacha20, otk_id =", otk_id, "data_size =", len(data))
		return r, e
	default:
		return nil, fmt.Errorf("unkown algo")
	}
}

// Wird verwendet um einen Verschlüsselten Datensatz mit dem Privaten Relay Schlüssel zu enschlüsseln
func (obj *Kernel) DecryptWithPrivateRelayKey(cipher_data []byte) ([]byte, error) {
	log.Println("Kernel: decrypting data with relay key. algo = chacha20, data_size =", len(cipher_data))
	return DecryptDataWithPrivateKey(obj._private_key, cipher_data)
}

// Wird verwendet um die Routen für ein Relay zu Laden
func (obj *Kernel) DumpsRoutesForRelayByConnection(conn RelayConnection) (bool, bool) {
	// Es wird versucht das Relay anhand der Verbindung aus dem Verbindungsmanager abzurufen
	relay, found, err := obj._connection_manager.GetRelayByConnection(conn)
	if err != nil {
		log.Println("Kernel: error by dumping routes for relays. error =", err.Error(), "connection =", conn.GetObjectId())
		return false, false
	}

	// Sollte kein Relay vorhanden sein, wird der Vorgang abgebrochen
	if !found {
		log.Println("Kernel: no relay for connection found. connection =", conn.GetObjectId())
	}

	// Es wird geprüft ob die Routen bereits Initalisiert wurden
	relay_active, routes_inited := obj._connection_manager.RelayAvailAndRoutesInited(relay)
	if !relay_active {
		if !conn.IsConnected() {
			log.Println("Kernel: error by dumping routes for relays, conenction was closed.", "connection =", conn.GetObjectId(), "relay =", relay.GetHexId())
			return false, false
		}
	}

	// Sollten die Routen bereits Initalisiert wurden sein, wird der vorgang abgebrochen
	if routes_inited {
		log.Println("Kernel: error by dumping routes for relays, routes always inited.", "connection =", conn.GetObjectId(), "relay =", relay.GetHexId())
		return true, false
	}

	// Es wird geprüft ob die Verbindung noch besteht, wenn nicht wird der Vorgang abgebrochen
	if !conn.IsConnected() {
		log.Println("Kernel: error by dumping routes for relays, conenction was closed.", "connection =", conn.GetObjectId(), "relay =", relay.GetHexId())
		return false, false
	}

	// Es werden alle Routen für diesen Relay aus der Routing Datenbank abgerufen
	routing_endpoints, err := obj._routing_table.FetchRoutesByRelay(relay)
	if err != nil {
		log.Println("Kernel: error by dumping routes for relays. error =", err.Error(), "connection =", conn.GetObjectId(), "relay =", relay.GetHexId())
		return false, false
	}

	// Es wird geprüft ob die Verbindung noch besteht, wenn nicht wird der Vorgang abgebrochen
	if !conn.IsConnected() {
		log.Println("Kernel: error by dumping routes for relays, conenction was closed.", "connection =", conn.GetObjectId(), "relay =", relay.GetHexId())
		return false, false
	}

	// Es wird wirderholt geprüft ob die Routen für diese Verbindungen bereits Initalisiert wurden
	relay_active, routes_inited = obj._connection_manager.RelayAvailAndRoutesInited(relay)
	if !relay_active {
		if !conn.IsConnected() {
			log.Println("Kernel: error by dumping routes for relays, conenction was closed.", "connection =", conn.GetObjectId(), "relay =", relay.GetHexId())
			return false, false
		}
	}

	// Sollten die Routen bereits Initalisiert wurden sein, wird der vorgang abgebrochen
	if routes_inited {
		log.Println("Kernel: error by dumping routes for relays, routes always inited.", "connection =", conn.GetObjectId(), "relay =", relay.GetHexId())
		return true, false
	}

	// Die Routen werden initalisiert
	if err := obj._connection_manager.InitRoutesForRelay(relay, routing_endpoints); err != nil {
		if conn.IsConnected() {
			log.Println("Kernel: error by dumping routes for relays. error =", err.Error(), "connection =", conn.GetObjectId(), "relay =", relay.GetHexId())
		} else {
			log.Println("Kernel: error by dumping routes for relays, conenction was closed.", "connection =", conn.GetObjectId(), "relay =", relay.GetHexId())
		}
		return false, false
	}

	// Log
	if len(routing_endpoints) > 0 {
		log.Println("Kernel: dumping relay routes.", "connection =", conn.GetObjectId(), "relay =", relay.GetHexId(), "total =", len(routing_endpoints))
		return true, true
	} else {
		log.Println("Kernel: dumping relay routes, no routes found.", "connection =", conn.GetObjectId(), "relay =", relay.GetHexId())
		return true, false
	}
}
