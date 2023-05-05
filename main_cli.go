//go:build cli

package main

import (
	"encoding/base32"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	apiclient "github.com/fluffelpuff/RoueX/api_client"
	"github.com/fluffelpuff/RoueX/static"
	"github.com/fluffelpuff/RoueX/utils"
	"github.com/olekukonko/tablewriter"
)

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func hexToBech32(h string) string {
	// Hex-String in Byte-Array umwandeln
	hexBytes, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}

	// Byte-Array in base32-String ohne Padding umwandeln
	base32Str := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hexBytes)
	return strings.ToLower(base32Str)
}

// Wird verwendet um alle Verfügbaren Relays abzurufen
func listRelays(list_all_relays bool) error {
	// Die API Verbindung wird aufgebaut
	api, err := apiclient.LoadAPI(static.GetFilePathFor(static.API_SOCKET))
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

	data := [][]string{}
	for i := range result {
		new_elm := []string{
			utils.ConvertHexStringToAddress(result[i].PublicKey),
			strconv.FormatUint(result[i].BandwithKBs, 10),
			strconv.FormatUint(result[i].TotalBytesSend, 10),
			strconv.FormatUint(result[i].TotalBytesRecived, 10),
			strconv.FormatUint(result[i].PingMS, 10),
			boolToYesNo(result[i].IsTrusted),
			boolToYesNo(result[i].IsConnected),
		}
		data = append(data, new_elm)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Public-Key / ID", "Bandwith (kb/s)", "TX (bytes)", "RX (bytes)", "Ping (ms)", "Trusted", "Connected"})

	for _, v := range data {
		table.Append(v)
	}

	table.Render()

	return nil
}

func main() {
	// Definiert alle Verwendeten werte
	var list_relays bool
	list_offline_relays := true

	// Definiert alle Parameter
	flag.BoolVar(&list_relays, "list-relays", false, "")
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
	} else {
		flag.Usage()
	}
}
