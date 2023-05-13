package kernel

import (
	"fmt"
	"log"

	apiclient "github.com/fluffelpuff/RoueX/api_client"
	"github.com/fxamacker/cbor"
)

// Stellt das Kernel API Interface dar
type Kf struct {
	_kernel     *Kernel
	_process_id string
}

// Ruft alle Verfügbren Relays ab
func (s *Kf) FetchRelays(_ apiclient.EmptyArg, reply *[]apiclient.ApiRelayEntry) error {
	// Es werden alle bekannten, verbunnden und vertrauten Relays abgerufen
	result, err := s._kernel.APIFetchAllRelays()
	if err != nil {
		return fmt.Errorf("FetchAllRelays: " + err.Error())
	}

	// Log
	log.Printf("KernelAPI-Session: fetched relays. connection = %s, total = %d\n", s._process_id, len(result))

	// Die Daten werden zurückgegeben
	*reply = result

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}

// Leitet API befehle an ein API Protocol weiter
func (s *Kf) PassCommandArgsToProtocol(args apiclient.CommandArgs, reply *[]byte) error {
	// Es wird geprüft ob das angegebene Protocol bekannt ist
	if !s._kernel.HasKernelProtocol(args.Id) {
		return fmt.Errorf("unkown protocol")
	}

	// Das Protokoll wird abgerufen
	protocol, err := s._kernel.GetKernelProtocolById(args.Id)
	if err != nil {
		return fmt.Errorf("unkown protocol")
	}

	// Der befehl und die Patameter werden an das Protokoll übergeben
	result, err := protocol.EnterCommandData(args.Method, args.Parms)
	if err != nil {
		return err
	}

	// Das Antwortpaket wird in Bytes umgewandelt und werden zurückgesendet
	cborData, err := cbor.Marshal(result, cbor.EncOptions{})
	if err != nil {
		panic(err)
	}

	// Die Daten werden zurückgegeben
	*reply = cborData

	// Der Vorgang wurde ohne Fehler beendet
	return nil
}

// Ruft alle Lokalen Adressen ab
func (s *Kf) FetchAllAddresses(_ apiclient.EmptyArg, reply *[]apiclient.ApiRelayEntry) error {
	// Es werden alle bekannten, verbunnden und vertrauten Relays abgerufen
	result, err := s._kernel.APIFetchAllRelays()
	if err != nil {
		return fmt.Errorf("FetchAllAddresses: " + err.Error())
	}

	// Die Daten werden zurückgegeben
	*reply = result

	// Der Vorgang wurde ohne Fehler durchgeführt
	return nil
}
