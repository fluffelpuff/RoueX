package kernel

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Wird ausgeführt um ausgehende Verbindungen zu verwalten
func manageOutboundConnection(k *Kernel, o RelayOutboundPair) {
	// Es wird ein neues Client Modul erstellt
	client_conn := *o.GetClientConnModule()

	// Diese Schleife wird solange ausgeführt bis der Kernel beendet wurde
	for k.IsRunning() {
		// Es wird eine ausgehende Verbindung estellt
		err := client_conn.ConnectTo(o.GetRelay().GetEndpoint(), o.GetRelay().GetPublicKey(), nil)
		if err != nil {
			log.Println("Outbound handler: " + err.Error())
			k.ServKernel(2500)
			continue
		}

		// Wird solange ausgeführt, bis die Verbindung geschlossen wurde
		client_conn.Serve()

		// Es wird 2 Sekunden geartet
		k.ServKernel(2000)
	}
}

// Verwaltet ausgehende Verbindungen
func outboundHandler(core *Kernel) {
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

// Händelt die System Events
func handleSystemEvents(core *Kernel) {
	// Erstelle einen Kanal, um Signale zu empfangen
	sigChan := make(chan os.Signal, 1)

	// Erlaube dem Kanal, die SIGINT- und SIGTERM-Signale zu empfangen
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Blockiere, bis ein Signal empfangen wird
	sig := <-sigChan

	// Es wird geprüft um welches Signal es sich handelt
	switch sig {
	case syscall.SIGINT:
		core.Shutdown()
		fmt.Println("God by")
	case syscall.SIGTERM:
		fmt.Println("TERM")
		break
	default:
		break
	}
}

// Hält den Mainthread am leben bis das Programm ein Close Event bekommt
func (obj *Kernel) Serve() error {
	// Der Kernel wird gestartet
	if err := obj.Start(); err != nil {
		return err
	}

	// Der Thread welcher System Events verarbeitet wird gestartet
	go handleSystemEvents(obj)

	// Dieser Thread wird ausgeführt um die Ausgehenden Verbindungen vorzubereiten
	go outboundHandler(obj)

	// Diese Schleife wird solange ausgefürth, solange der Kernel ausgeführt wird
	for range time.Tick(1 * time.Millisecond) {
		if obj._is_closed() {
			break
		}
	}

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}
