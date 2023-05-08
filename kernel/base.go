package kernel

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/fluffelpuff/RoueX/static"
	"github.com/fluffelpuff/RoueX/utils"
)

// Stellt das Kernel Objekt dar
type Kernel struct {
	_external_modules_path string
	_os_path_trimmer       string
	_socket_path           string
	_kernel_id             string
	_is_running            bool
	_lock                  *sync.Mutex
	_routing_table         *RoutingManager
	_trusted_relays        *TrustedRelays
	_server_modules        []*ServerModule
	_client_modules        []*ClientModule
	_connection_manager    RelayConnectionRoutingTable
	_firewall              *Firewall
	_private_key           *btcec.PrivateKey
	_api_interfaces        []*KernelAPI
	_temp_key_pairs        map[string]*btcec.PrivateKey
	_temp_ecdh_keys        map[string][]byte
	_adr_layer_feps        []*kernel_package_type_function_entry
}

// Wird unter UNIX, Linux oder Windows verwendet zum aufräumen
func (obj *Kernel) CleanUp() error {
	log.Println("Clearing kernel...")
	return nil
}

// Wird verwendet um zu Warten
func (obj *Kernel) Waiter(wms uint64) {
	ticks := 0
	for obj.IsRunning() {
		if ticks >= int(wms) {
			return
		}
		ticks++
		time.Sleep(1 * time.Millisecond)
	}
}

// Gibt an ob der Kernel ausgeführt wird
func (obj *Kernel) IsRunning() bool {
	obj._lock.Lock()
	is_running := obj._is_running
	obj._lock.Unlock()
	return is_running
}

// Registriert eine API Schnitstellt
func (obj *Kernel) RegisterAPIInterface(api_interace *KernelAPI) error {
	obj._lock.Lock()
	if err := api_interace._register_kernel(obj); err != nil {
		if err != nil {
			obj._lock.Unlock()
			return fmt.Errorf("RegisterAPIInterface: " + err.Error())
		}
	}
	obj._api_interfaces = append(obj._api_interfaces, api_interace)
	obj._lock.Unlock()
	return nil
}

// Gibt an ob es sich um eine Lokale Adresse handelt
func (obj *Kernel) IsLocallyAddress(pubkey btcec.PublicKey) bool {
	return false
}

// Erstellt einen UNIX Kernel
func CreateUnixKernel(priv_key *btcec.PrivateKey) (*Kernel, error) {
	// Log
	fmt.Println("Creating new RoueX UNIX Kernel...")

	// Es wird eine Liste mit allen Vertrauten Relays abgerufen
	trusted_relays_obj, err := loadTrustedRelaysTable(static.GetFilePathFor(static.TRUSTED_RELAYS))
	if err != nil {
		log.Fatal("listen error:", err)
		return nil, err
	}

	// Es wird versucht die Routing Tabelle zu laden
	routing_table_obj, err := loadRoutingManager(static.GetFilePathFor(static.ROUTING_TABLE))
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

	// Die Kernel API wird gestartet
	kernel_api, err := newKernelAPI(static.GetFilePathFor(static.API_SOCKET), static.GetFilePathFor(static.CHANNEL_PATH))
	if err != nil {
		panic(err)
	}

	// Die KernelID wird estellt
	k_id := utils.RandStringRunes(16)

	// Die Verbindungsverwaltung wird erstellt
	conn_manager := newRelayConnectionRoutingTable()

	// Erstellt das Kernel Objekt
	new_kernel := Kernel{
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
		_api_interfaces:        make([]*KernelAPI, 0),
		_external_modules_path: static.OSX_EXTERNAL_MODULES,
		_adr_layer_feps:        []*kernel_package_type_function_entry{},
	}

	// Die API Schnitstelle wird im Kernel Registriert
	if err := new_kernel.RegisterAPIInterface(kernel_api); err != nil {
		panic(err)
	}

	// Gibt das Kernelobjekt ohne Fehler zurück
	log.Println("Kernel: new unix kernel created. id =", k_id)
	return &new_kernel, nil
}
