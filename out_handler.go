package main

import (
	"log"

	"github.com/fluffelpuff/RoueX/kernel"
)

// Wird ausgeführt um ausgehende Verbindungen zu verwalten
func manageOutboundConnection(k *kernel.Kernel, o kernel.RelayOutboundPair) {
	// Es wird ein neues Client Modul erstellt
	client_conn := *o.GetClientConnModule()

	// Diese Schleife wird solange ausgeführt bis der Kernel beendet wurde
	for k.IsRunning() {
		// Es wird eine ausgehende Verbindung estellt
		err := client_conn.ConnectTo(o.GetRelay().GetEndpoint(), o.GetRelay().GetPublicKey(), nil)
		if err != nil {
			log.Println("Outbound handler: " + err.Error())
			k.Waiter(2500)
			continue
		}

		// Wird solange ausgeführt, bis die Verbindung geschlossen wurde
		client_conn.Serve()
		log.Println("Outbound handler: outbound relay connection closed. id =", client_conn.GetObjectId(), "reconnection in 2 seconds.")

		// Es wird 2 Sekunden geartet
		k.Waiter(2000)
	}
}

// Verwaltet ausgehende Verbindungen
func outboundHandler(core *kernel.Kernel) {
	// Es wird versucht alle Ausgehenden Relays abzurufen
	rt, err := core.ListOutboundTrustedAvaileRelays()
	if err != nil {
		panic(err)
	}

	// Es wird versucht mit jedem Relay eine verbindung aufzubauen
	for _, o := range rt {
		go manageOutboundConnection(core, o)
	}
}
