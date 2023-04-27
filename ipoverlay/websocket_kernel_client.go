package ipoverlay

import (
	"fmt"

	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/gorilla/websocket"
)

type WebsocketKernelClient struct {
	_obj_id string
}

// Registriert den Kernel im Modul
func (obj *WebsocketKernelClient) RegisterKernel(kernel *kernel.Kernel) error {
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
func (obj *WebsocketKernelClient) ConnectTo(url string, pub_key string) error {
	// Log
	fmt.Printf("Trying to establish a websocket connection to %s...\n", url)

	// Es wird versucht eine Websocket verbindung aufzubauen
	_, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("ConnectTo: 1: " + err.Error())
	}

	// Log
	fmt.Printf("The websocket base connection to %s has been established.\n", url)

	// Der Gegenseite wird nun der eigene Öffentliche Schlüssel, die Aktuelle Uhrzeit sowie
	return nil
}

// Gibt die Aktuelle ObjektID aus
func (obj *WebsocketKernelClient) GetObjectId() string {
	return obj._obj_id
}

// Beendet das Module, verhindert das weitere verwenden
func (obj *WebsocketKernelClient) Shutdown() {
	return
}

// Erstellt ein neues Websocket Client Modul
func NewWebsocketClient() *WebsocketKernelClient {
	rand_id := kernel.RandStringRunes(16)
	return &WebsocketKernelClient{_obj_id: rand_id}
}
