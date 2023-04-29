package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fluffelpuff/RoueX/ipoverlay"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fluffelpuff/RoueX/keystore"
)

func main() {
	// Die Einstellungen werden geladen
	if err := loadConfigs(); err != nil {
		panic(err)
	}

	// Es wird versucht den Privaten Schlüssel zu laden
	pub_key, priv_key, err := keystore.LoadPrivateKeyFromKeyStore()
	if err != nil {
		panic(err)
	}

	// Log
	fmt.Println("Public relay key:", hex.EncodeToString(pub_key.SerializeCompressed()))

	// Das Passende Systemkernel wird erstellt
	kernel_object, err := kernel.CreateOSXKernel(priv_key)
	if err != nil {
		panic(err)
	}

	// Der Kernel Räumt die Tabellen auf
	if err := kernel_object.CleanUp(); err != nil {
		panic(err)
	}

	// Es wird ein Lokaler Websocket Server erezugt
	local_ws, err := ipoverlay.CreateNewLocalWebsocketServerEP("", 9381)
	if err != nil {
		panic(err)
	}

	// Der Websocket Client wird ersteltlt
	ws_client_module := ipoverlay.NewWebsocketClient()

	// Der Lokale Websocket Server Endpunkt wird hinzugefügt
	if err := kernel_object.RegisterServerModule(local_ws); err != nil {
		panic(err)
	}

	// Das Websocket Client Protokoll wird registriert
	if err := kernel_object.RegisterClientModule(ws_client_module); err != nil {
		panic(err)
	}

	// Es wird nach allen Sub Modulen gesucht
	if err := kernel_object.LoadExternalKernelModules(); err != nil {
		panic(err)
	}

	// Der Kernel wird gestartet
	if err := kernel_object.Start(); err != nil {
		panic(err)
	}

	// Der Websocket Server wird gestartet
	if err := local_ws.Start(); err != nil {
		panic(err)
	}

	// Die Externen Kernel Module werden gestartet
	if err := kernel_object.StartExternalKernelModules(); err != nil {
		panic(err)
	}

	// Diese Funktion wird ausgeführt sobald STRG-C gedrückt wird
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		kernel_object.Shutdown()
	}()

	// Dieser Thread wird ausgeführt um die Ausgehenden Verbindungen vorzubereiten
	go outboundHandler(kernel_object)

	// Diese Schleife wird solange ausgefürth, solange der Kernel ausgeführt wird
	for kernel_object.IsRunning() {
		time.Sleep(1 * time.Millisecond)
	}
}
