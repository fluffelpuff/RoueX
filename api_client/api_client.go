package apiclient

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"

	"github.com/fluffelpuff/RoueX/static"
)

type APIClient struct {
	_client         *rpc.Client
	_channel_client net.Conn
	_lock           *sync.Mutex
}

// Ruft alle Vefügbaren Relays ab
func (obj *APIClient) FetchAllRelays() ([]ApiRelayEntry, error) {
	// Aufruf der Methode "SomeMethod" auf dem RPC-Server
	var reply []ApiRelayEntry
	err := obj._client.Call("Kf.FetchRelays", EmptyArg{}, &reply)
	if err != nil {
		log.Fatal("Fehler beim Aufruf der Methode 'SomeMethod':", err)
	}
	return reply, nil
}

// Schließt die Verbindung
func (obj *APIClient) Close() {
	obj._lock.Lock()
	obj._client.Close()
	obj._channel_client.Close()
	obj._lock.Unlock()
}

// Erstellt eine neue API
func LoadAPI() (*APIClient, error) {
	// Der Pfad für den API Socket wird abgerufen
	rpc_path, channel_path := static.GetFilePathFor(static.API_SOCKET), static.GetFilePathFor(static.CHANNEL_PATH)

	// Es wird versucht eine Socket verbindung aufzubauen
	rpc_client, err := rpc.Dial("unix", rpc_path)
	if err != nil {
		return nil, fmt.Errorf("error by connection to rpc service" + err.Error())
	}

	// Die Channel Verbindung wird aufgebaut
	channel_conn, err := net.Dial("unix", channel_path)
	if err != nil {
		return nil, fmt.Errorf("error by connection to channel service" + err.Error())
	}

	// Das Rückgabe Objekt wird erstellt
	return &APIClient{_client: rpc_client, _channel_client: channel_conn, _lock: &sync.Mutex{}}, nil
}
