package ipoverlay

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type WebsocketKernelServerEP struct {
	_kernel *kernel.Kernel
	_obj_id string
}

// Registriert den Kernel im Module
func (obj *WebsocketKernelServerEP) RegisterKernel(k *kernel.Kernel) error {
	log.Printf("Websocket Server EndPoint registrated on kernel %s\n", k.GetKernelID())
	obj._kernel = k
	return nil
}

// Wird verwendet um den Serversocket herunterzufahren
func (obj *WebsocketKernelServerEP) Shutdown() {
	log.Printf("Websocket Server EndPoint shutingdown...\n")
}

// Startet den eigentlichen Server
func (obj *WebsocketKernelServerEP) Start() error {
	log.Printf("New Websocket Server EndPoint started\n")
	return nil
}

// Gibt das Aktuelle Protokoll aus
func (obj *WebsocketKernelServerEP) GetProtocol() string {
	return "ws"
}

// Gibt die Aktuelle Objekt ID aus
func (obj *WebsocketKernelServerEP) GetObjectId() string {
	return obj._obj_id
}

// Gibt an ob der Server bereits gestartet wurde
func (obj *WebsocketKernelServerEP) IsRunning() bool {
	return false
}

// Upgradet die HTTP Verbindung und erstellt eine Client Sitzung daraus
func (obj *WebsocketKernelServerEP) upgradeHTTPConnAndRegister(w http.ResponseWriter, r *http.Request) {
	// Die Verbindung wird zu einer Websocket Verbindung geupgradet zu einer Websocket verbindung
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Es wird auf das eintreffende Paket gewartet
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("received message: %s\n", message)

	err = conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Println(err)
		return
	}
}

// Erstellt einen neuen Lokalen Websocket Server
func CreateNewLocalWebsocketServerEP(ip_adr string, port uint64) (*WebsocketKernelServerEP, error) {
	// Die Einmalige ObjektID wird erstellt
	rand_id := kernel.RandStringRunes(16)

	// Das Objekt wird vorbereitet
	result_obj := &WebsocketKernelServerEP{_obj_id: rand_id}

	// Es wird eine neuer Thread gestartet, innerhalb dieses Threads wird der HTTP Server ausgeführt
	go func() {
		server := &http.Server{
			Addr:    ":" + strconv.Itoa(int(port)),
			Handler: http.HandlerFunc(result_obj.upgradeHTTPConnAndRegister),
		}

		// Der Server wird ausgeführt
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Es wird eine zufälliger Objekt ID erstellt
	log.Printf("New Websocket Server EndPoint on %s and port %d created\n", ip_adr, port)
	return result_obj, nil
}
