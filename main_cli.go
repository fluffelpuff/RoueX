//go:build cli

package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
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
func pingRelayAddress(relay_address string) {
	// Es wird versucht die Adresse zu dekodieren
	decoded_address, err := utils.ConvertAddressToPublicKey(relay_address)
	if err != nil {
		fmt.Println("PingRelayAddress: " + err.Error())
		return
	}

	// Die API Verbindung wird aufgebaut
	api, err := apiclient.LoadAPI()
	if err != nil {
		fmt.Println("PingRelayAddress: " + err.Error())
		return
	}

	// Schließt die Verbindug am ende
	defer api.Close()

	// Wird ausgeführt bis der User sie abbricht
	for {
		rt, err := api.PingAddress(decoded_address, uint16(12000))
		if err != nil {
			panic(err)
		}
		if rt == -1 {
			fmt.Println("Request time out")
		} else {
			fmt.Printf("Answer from %s: Time %d ms\n", relay_address, rt)
		}
		time.Sleep(1 * time.Second)
	}
}

// Wandelt einen Hex PublicKey in eine Adresse um
func convertoToAddress(address_hx_str string) error {
	// Es wird versucht den Öffentlichen Schlüssel einzulesne
	decodec, err := hex.DecodeString(address_hx_str)
	if err != nil {
		return err
	}
	readed_public_key, err := btcec.ParsePubKey(decodec)
	if err != nil {
		return err
	}

	// Der Öffentliche Schlüssel wird in die Adresse umgewandelt
	convadr := utils.ConvertPublicKeyToAddress(readed_public_key)
	fmt.Println(convadr)

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

func main() {
	// Definiert alle Verwendeten werte
	var convertPublicKeyToAddress string
	var list_relays bool
	var pingArg string
	list_offline_relays := true

	// Definiert alle Parameter
	flag.BoolVar(&list_relays, "list-relays", false, "")
	flag.StringVar(&pingArg, "ping", "", "description of ping flag")
	flag.BoolVar(&list_offline_relays, "all", false, "A boolean flag")
	flag.StringVar(&convertPublicKeyToAddress, "convert-to-address", "", "description of ping flag")

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
		pingRelayAddress(pingArg)
	} else if len(convertPublicKeyToAddress) != 0 {
		if err := convertoToAddress(convertPublicKeyToAddress); err != nil {
			panic(err)
		}
	} else {
		flag.Usage()
	}
}
