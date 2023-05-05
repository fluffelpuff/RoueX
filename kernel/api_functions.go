package kernel

import (
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
			result_list = append(result_list, apiclient.ApiRelayEntry{
				Id:                trusted_relays[i]._hexed_id,
				IsTrusted:         trusted_relays[i].IsTrusted(),
				PublicKey:         trusted_relays[i].GetPublicKeyHexString(),
				IsConnected:       meta_data.IsConnected,
				BandwithKBs:       meta_data.BandwithKBs,
				TotalConnections:  meta_data.TotalConnections,
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
