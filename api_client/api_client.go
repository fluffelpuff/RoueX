package apiclient

import (
	"fmt"
	"net/rpc"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/static"
)

// Gibt den Status eines Ping Vorganges an
type ping_state uint8

// Definiert alle verfügbaren Ping Vorgänge
const (
	ABORTED          = ping_state(0)
	RESPONDED        = ping_state(1)
	CLOSED_BY_KERNEL = ping_state(2)
	TIMEOUT          = ping_state(3)
	INTERNAL_ERROR   = ping_state(4)
)

// Stellt eine API Verbindung dar
type APIClient struct {
	_client *rpc.Client
	_lock   *sync.Mutex
}

// Ruft alle Vefügbaren Relays ab
func (obj *APIClient) FetchAllRelays() ([]ApiRelayEntry, error) {
	// Aufruf der Methode "FetchAllRelays" auf dem RPC-Server
	var reply []ApiRelayEntry
	err := obj._client.Call("Kf.FetchRelays", EmptyArg{}, &reply)
	if err != nil {
		return nil, fmt.Errorf("FetchAllRelays: Fehler beim Aufruf der Methode 'FetchAllRelays'" + err.Error())
	}

	// Die Daten werden zurückgegeben
	return reply, nil
}

// Wird verwendet um einen Ping vorgang durchzuführen
func (obj *APIClient) PingAddress(adr *btcec.PublicKey, timeout uint16) (int64, error) {
	// Aufruf der Methode "PingAddress" auf dem RPC-Server
	var reply map[string]interface{}
	err := obj._client.Call("Kf.PassCommandArgsToProtocol", CommandArgs{Id: PING_PROTOCOL, Method: "ping_address", Parms: [][]byte{adr.SerializeCompressed()}}, &reply)
	if err != nil {
		return -1, fmt.Errorf("PingAddress: " + err.Error())
	}

	// Es wird geprüft ob der state Wert vorhanden ist
	if val, ok := reply["state"]; !ok && val == nil {
		panic("internal error -1")
	}

	// Es wird versucht den Status zurückzuwandeln
	state, ok := reply["state"].(uint8)
	if !ok {
		return -1, fmt.Errorf("invalid state type")
	}

	// Der Aktuelle Status wird ermittelt
	switch state {
	case uint8(ABORTED):
		return -1, fmt.Errorf("aborted")
	case uint8(RESPONDED):
		// Es wird geprüft ob der Timewert vorhadnen ist
		if val, ok := reply["state"]; !ok && val == nil {
			panic("internal error -1")
		}

		// Es wird versucht die Zeit einzulesen
		ttime, ok := reply["ttime"].(uint64)
		if !ok {
			return -1, fmt.Errorf("invalid state type")
		}

		// Die Daten werden zurückgegeben
		return int64(ttime), nil
	case uint8(CLOSED_BY_KERNEL):
		return -1, fmt.Errorf("api is shutingdown")
	case uint8(TIMEOUT):
		return -1, nil
	default:
		return -1, fmt.Errorf("internal error")
	}
}

// Schließt die Verbindung
func (obj *APIClient) Close() {
	obj._lock.Lock()
	obj._client.Close()
	obj._lock.Unlock()
}

// Erstellt eine neue API
func LoadAPI() (*APIClient, error) {
	// Der Pfad für den API Socket wird abgerufen
	rpc_path := static.GetFilePathFor(static.API_SOCKET)

	// Es wird versucht eine Socket verbindung aufzubauen
	rpc_client, err := rpc.Dial("unix", rpc_path)
	if err != nil {
		return nil, fmt.Errorf("error by connection to rpc service" + err.Error())
	}

	// Das Rückgabe Objekt wird erstellt
	return &APIClient{_client: rpc_client, _lock: &sync.Mutex{}}, nil
}
