package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fluffelpuff/RoueX/ipoverlay"
	"github.com/fluffelpuff/RoueX/kernel"
)

func outboundHandler(core *kernel.Kernel) {
	rt, err := core.ListOutboundAvaileRelays()
	if err != nil {
		panic(err)
	}

	for _, o := range rt {
		client_conn := *o.GetClientConnModule()
		err := client_conn.ConnectTo(o.GetRelay().GetEndpoint(), o.GetRelay().GetPublicKey())
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(o.GetRelay().GetProtocol(), o.GetRelay().GetEndpoint())
	}
}

func main() {
	// Das Passende Systemkernel wird erstellt
	kernel_object, err := kernel.CreateOSXKernel("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
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
