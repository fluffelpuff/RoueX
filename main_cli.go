//go:build cli

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	apiclient "github.com/fluffelpuff/RoueX/api_client"
	"github.com/fluffelpuff/RoueX/kernel"
	"github.com/fluffelpuff/RoueX/utils"
)

// Wird verwendet um alle Verfügbaren Relays abzurufen
func listRelays(list_all_relays bool) error {
	// Die API Verbindung wird aufgebaut
	api, err := apiclient.LoadAPI()
	if err != nil {
		return err
	}

	// Schließt die Verbindug am ende
	defer api.Close()

	// Es wird geprüf ob auch alle Offline Relays abgerufen werden sollen
	result, err := api.FetchAllRelays()
	if err != nil {
		panic(err)
	}

	// Sollten keine Relays vorhanden sein, wird eine Warnung ausgegeben
	if len(result) < 1 {
		fmt.Printf("No relays available, you are offline.\n")
		return nil
	}

	// Print interface details
	for _, iface := range result {
		// Speichert alle Optionen ab
		var options []string

		// Es wird geprüft ob eine Verbindung besteht
		if iface.IsConnected {
			options = append(options, "CONNECTED")
		} else {
			options = append(options, "DISCONNECTED")
		}

		// Prüft ob der Verbindung vertraut wird
		if iface.IsTrusted {
			options = append(options, "TRUSTED")
		} else {
			options = append(options, "UNTRUSTED")
		}

		// Erstellt den Ausagbe String aus den Optionen
		joinedFlags := strings.Join(options, ",")

		// Erzeugt die ausgabe
		fmt.Printf("%s: <%s>\n", iface.Id, joinedFlags)
		fmt.Printf("\trealy pkey: %s\n", utils.ConvertHexStringToAddress(iface.PublicKey))
		for _, connection := range iface.Connections {
			if kernel.ConnectionIoType(connection.InboundOutbound) == kernel.INBOUND {
				fmt.Printf("\tin: spkey = %s, protocol = %s, ping = %d ms, tx = %d bytes, rx = %d bytes\n", connection.Id, connection.Protocol, connection.Ping, connection.TxBytes, connection.RxBytes)
			} else if kernel.ConnectionIoType(connection.InboundOutbound) == kernel.OUTBOUND {
				fmt.Printf("\tout: spkey = %s, protocol = %s, ping = %d ms, tx = %d bytes, rx = %d bytes\n", connection.Id, connection.Protocol, connection.Ping, connection.TxBytes, connection.RxBytes)
			} else {
				continue
			}
		}
		fmt.Printf("\ttotal bytes recived: %d\n", iface.TotalBytesRecived)
		fmt.Printf("\ttotal bytes send: %d\n", iface.TotalBytesSend)
		fmt.Printf("\ttotal connections: %d\n", iface.TotalConnections)
		fmt.Printf("\tping (ms): %d\n", iface.PingMS)
	}

	// Der Vorgang wurde ohne fehler durchgeführt
	return nil
}

// Es wird ein Ping vorgang gestartet
func pingRelayAddress(relay_address string) error {
	// Die API Verbindung wird aufgebaut
	api, err := apiclient.LoadAPI()
	if err != nil {
		return err
	}

	// Schließt die Verbindug am ende
	defer api.Close()

	// Es wird geprüf ob auch alle Offline Relays abgerufen werden sollen
	_, err = api.FetchAllRelays()
	if err != nil {
		panic(err)
	}

	// Der Vorgang wurde ohne fehler durchgeführt
	return nil
}

func main() {
	// Definiert alle Verwendeten werte
	var pingArg string
	var list_relays bool
	list_offline_relays := true

	// Definiert alle Parameter
	flag.BoolVar(&list_relays, "list-relays", false, "")
	flag.StringVar(&pingArg, "ping", "", "description of ping flag")
	flag.BoolVar(&list_offline_relays, "all", false, "A boolean flag")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\t-list-relays: Liste Relays auf\n")
		fmt.Fprintf(os.Stderr, "\t-list-connections: Liste Verbindungen auf\n")
	}

	// Parst alle Parameter
	flag.Parse()

	// Es wird geprüft welche Option aktiviert wurde
	if list_relays {
		if err := listRelays(list_offline_relays); err != nil {
			panic(err)
		}
	} else if len(pingArg) != 0 {
		if err := pingRelayAddress(pingArg); err != nil {
			panic(err)
		}
	} else {
		flag.Usage()
	}
}
