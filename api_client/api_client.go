package apiclient

import (
	"fmt"
	"net/rpc"
	"sync"

	"github.com/fluffelpuff/RoueX/static"
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
