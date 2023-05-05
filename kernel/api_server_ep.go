package kernel

import (
	"fmt"

	apiclient "github.com/fluffelpuff/RoueX/api_client"
)

// Stellt das Kernel API Interface dar
type Kf struct {
	_kernel     *Kernel
	_kernel_api *KernelAPI
}

// Ruft alle Verfügbren Relays ab
func (s *Kf) FetchRelays(_ apiclient.EmptyArg, reply *[]apiclient.ApiRelayEntry) error {
	// Es werden alle bekannten, verbunnden und vertrauten Relays abgerufen
	result, err := s._kernel.APIFetchAllRelays()
	if err != nil {
		return fmt.Errorf("FetchAllRelays: " + err.Error())
	}

	// Die Daten werden zurückgegeben
	*reply = result

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}
