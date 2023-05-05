package main

import (
	"fmt"

	"github.com/fluffelpuff/RoueX/kernel"
)

// Verwaltet ausgehende Verbindungen
func outboundHandler(core *kernel.Kernel) {
	// Es wird versucht alle Ausgehenden Relays abzurufen
	rt, err := core.ListOutboundTrustedAvaileRelays()
	if err != nil {
		panic(err)
	}

	// Es wird versucht mit jedem Relay eine verbindung aufzubauen
	for _, o := range rt {
		client_conn := *o.GetClientConnModule()
		err := client_conn.ConnectTo(o.GetRelay().GetEndpoint(), o.GetRelay().GetPublicKey(), nil)
		if err != nil {
			fmt.Println(err)
		}
	}
}
