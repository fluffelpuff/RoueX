package apiclient

import (
	"fmt"
	"log"
	"net/rpc"
)

type APIClient struct {
	_client *rpc.Client
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
	obj._client.Close()
}

// Erstellt eine neue API
func LoadAPI(path string) (*APIClient, error) {
	// Es wird versucht eine Socket verbindung aufzubauen
	client, err := rpc.Dial("unix", path)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Herstellen der Verbindung zum RPC-Server:" + err.Error())
	}

	// Das Rückgabe Objekt wird erstellt
	return &APIClient{_client: client}, nil
}
