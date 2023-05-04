package kernel

import (
	"fmt"
	"log"
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
	_routing_table         *RoutingTable
	_trusted_relays        *TrustedRelays
	_server_modules        []*ServerModule
	_client_modules        []*ClientModule
	_connection_manager    ConnectionManager
	_firewall              *Firewall
	_private_key           *btcec.PrivateKey
	_api_interfaces        []*KernelAPI
	_temp_key_pairs        map[string]*btcec.PrivateKey
	_temp_ecdh_keys        map[string][]byte
}

// Wird unter OSX, Linux oder Windows verwendet zum aufräumen
func (obj *Kernel) CleanUp() error {
	log.Println("Clearing kernel...")
	return nil
}

// Wird ausgeführt wenn das Programm als Dienst ausgeführt wird
func (obj *Kernel) Start() error {
	obj._lock.Lock()
	for i := range obj._api_interfaces {
		if err := obj._api_interfaces[i]._start_by_kernel(); err != nil {
			obj._lock.Unlock()
			return err
		}
	}
	obj._is_running = true
	obj._lock.Unlock()
	return nil
}

// Wird ausgeführt um den Kernel zu beenden
func (obj *Kernel) Shutdown() {
	obj._lock.Lock()
	if obj._is_running {
		// Die Internen Dienste werden beendet
		var vat ServerModule
		for _, item := range obj._server_modules {
			if item == nil {
				continue
			}
			vat = *item
			vat.Shutdown()
		}

		// Die API Schnitstellen werden geschlossen
		for i := range obj._api_interfaces {
			obj._api_interfaces[i]._close_by_kernel()
		}

		// Es werden alle Verbindungen geschlossen
		obj._connection_manager.ShutdownByKernel()

		// Die Datenbanken werden geschlossen
		obj._routing_table.Shutdown()
		obj._trusted_relays.Shutdown()

		// Es wird Signalisiert dass der Kernel nicht mehr läuft
		obj._is_running = false

		// Der Threadlock wird freigegeben
		obj._lock.Unlock()

		// Log
		log.Println("Kernel shutdown...")
		return
	}
	obj._lock.Unlock()
}

// Gibt an ob der Kernel ausgeführt wird
func (obj *Kernel) IsRunning() bool {
	var is_running bool
	obj._lock.Lock()
	is_running = obj._is_running
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

	// Die Kernel API wird gestartet
	kernel_api, err := newKernelAPI(static.GetFilePathFor(static.API_SOCKET))
	if err != nil {
		panic(err)
	}

	// Log
	fmt.Println("RoueX OSX Kernel API Unix-Socket created...")

	// Die KernelID wird estellt
	k_id := RandStringRunes(16)

	// Die Verbindungsverwaltung wird erstellt
	conn_manager := newConnectionManager()

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
	}

	// Die API Schnitstelle wird im Kernel Registriert
	if err := new_kernel.RegisterAPIInterface(kernel_api); err != nil {
		panic(err)
	}

	// Gibt das Kernelobjekt ohne Fehler zurück
	fmt.Printf("New OSX RoueX Kernel '%s' created...\n", k_id)
	return &new_kernel, nil
}
