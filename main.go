//go:build !cli

package main

import (
	"encoding/hex"
	"fmt"

	"github.com/fluffelpuff/RoueX/ipoverlay"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fluffelpuff/RoueX/keystore"
	protocols "github.com/fluffelpuff/RoueX/protocols"
	"github.com/fluffelpuff/RoueX/static"
)

func main() {
	// Der Banner wird angezeigt
	fmt.Print(static.WELCOME_BANNER)

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
	kernel_object, err := kernel.CreateUnixKernel(priv_key)
	if err != nil {
		panic(err)
	}

	// Das Ping Pong Layer 2 Protkoll wird Registriert
	layer_two_ping_pong := protocols.NEW_ROUEX_PING_PONG_PROTOCOL_HANDLER()
	if err := kernel_object.RegisterNewKernelTypeProtocol(0, layer_two_ping_pong); err != nil {
		panic(err)
	}

	// Es wird ein Lokaler Websocket Server erezugt
	local_ws, err := ipoverlay.CreateNewLocalWebsocketServerEP("", static.WS_PORT)
	if err != nil {
		panic(err)
	}

	// Der Websocket Client Hadnler wird ersteltlt
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

	// Der Kernel wird ausgeführt
	if err := kernel_object.Serve(); err != nil {
		fmt.Println(err)
	}

	// Gibt den God by text aus
	fmt.Println("God by")
}
