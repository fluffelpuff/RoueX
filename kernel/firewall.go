package kernel

import "fmt"

type Firewall struct {
}

func loadFirewallTable(path string) (*Firewall, error) {
	// Es wird geprüft ob die Datei vorhanden ist, wenn nicht wird sie erzeugt mit allen Standard Tabellen

	fmt.Printf("Loading Firewall database from %s\n", path)
	return nil, nil
}
