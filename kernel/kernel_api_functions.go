package kernel

import (
	"fmt"

	apiclient "github.com/fluffelpuff/RoueX/api_client"
)

// Ruft alle Relays ab
func (obj *Kernel) APIFetchAllRelays() ([]apiclient.ApiRelayEntry, error) {
	// Es werden alle Trusted Relays abgerufen
	trusted_relays := obj._trusted_relays.GetAllRelays()

	// Die Rückgabe Liste wird erstellt
	result_list := make([]apiclient.ApiRelayEntry, 0)
	for i := range trusted_relays {
		// Es wird versucht alle Meta Daten der Verbindung aus dem Verbindungs Manager abzurufen
		meta_data, err := obj._connection_manager.GetAllMetaInformationsOfRelayConnections(trusted_relays[i])
		if err != nil {
			return nil, err
		}

		// Es wird geprüft ob der Relay verbunden ist
		if meta_data != nil {
			// Die Verbindungs Meta Daten werden vorbereitet
			recons := make([]apiclient.ApiRelayConnection, 0)
			for i := range meta_data.Connections {
				recons = append(recons, apiclient.ApiRelayConnection{
					Id:              meta_data.Connections[i].Id,
					SessionPkey:     meta_data.Connections[i].SessionPKey,
					Protocol:        meta_data.Connections[i].Protocol,
					InboundOutbound: meta_data.Connections[i].InboundOutbound,
					TxBytes:         meta_data.Connections[i].TxBytes,
					RxBytes:         meta_data.Connections[i].RxBytes,
					Ping:            meta_data.Connections[i].Ping,
				})
			}

			// Die Daten werden hinzugefügt
			result_list = append(result_list, apiclient.ApiRelayEntry{
				Id:                trusted_relays[i]._hexed_id,
				IsTrusted:         trusted_relays[i].IsTrusted(),
				PublicKey:         trusted_relays[i].GetPublicKeyHexString(),
				Connections:       recons,
				IsConnected:       meta_data.IsConnected,
				TotalConnections:  uint64(len(meta_data.Connections)),
				TotalBytesSend:    meta_data.TotalWrited,
				TotalBytesRecived: meta_data.TotalReaded,
				PingMS:            meta_data.PingMS,
			})
		} else {
			result_list = append(result_list, apiclient.ApiRelayEntry{
				Id:                trusted_relays[i]._hexed_id,
				PublicKey:         trusted_relays[i].GetPublicKeyHexString(),
				IsTrusted:         trusted_relays[i].IsTrusted(),
				IsConnected:       false,
				BandwithKBs:       0,
				TotalConnections:  0,
				TotalBytesSend:    0,
				TotalBytesRecived: 0,
				PingMS:            0,
			})
		}
	}

	// Die Daten werden zurückgegebn
	return result_list, nil
}

// Gibt an ob ein bestimmtes Protocol vorhanden ist
func (obj *Kernel) HasKernelProtocol(protocol_type uint8) bool {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob die
	var found_value *KernelTypeProtocol
	for i := range obj._protocols {
		if obj._protocols[i].Tpe == protocol_type {
			found_value = &obj._protocols[i].Ptf
			break
		}
	}

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Der Status wird wieder zurückgegeben
	return found_value != nil
}

// Gibt das Kernel Protokoll zurück
func (obj *Kernel) GetKernelProtocolById(protocol_type uint8) (KernelTypeProtocol, error) {
	// Der Threadlock wird ausgeführt
	obj._lock.Lock()

	// Es wird geprüft ob die
	var found_value *KernelTypeProtocol
	for i := range obj._protocols {
		if obj._protocols[i].Tpe == protocol_type {
			found_value = &obj._protocols[i].Ptf
			break
		}
	}

	// Der Threadlock wird freigegeben
	obj._lock.Unlock()

	// Sollte kein Protokoll gefunden wurden sein, wird der Vorgang abgebrochen
	if found_value == nil {
		return nil, fmt.Errorf("unkown protocol")
	}

	// Die Daten werden ohne Fehler zurückgegeben
	return *found_value, nil
}
