package ipoverlay

import (
	"github.com/fluffelpuff/RoueX/static"
	"github.com/fxamacker/cbor"
)

// Flag Abbildung
type WSPackageFlag struct {
	// Gibt den Verwendeten Flag mit 1 Byte an
	Flag []byte `cbor:"1,keyasint"`

	// Gibt den Optionalen Flag wert an (maximal 96 bytes)
	Value []byte `cbor:"2,keyasint"`
}

// Erstellt ein neues Flgag Objekt
func NewWSPackageFlag(flag []byte, value []byte) (WSPackageFlag, error) {
	return WSPackageFlag{Flag: flag, Value: value}, nil
}

// Sitzungsinitialisierung
type EncryptedClientHelloPackage struct {
	// Der Öffentliche Server Schlüssel
	PublicServerKey []byte `cbor:"3,keyasint"`

	// Der Öffentliche Client Schlüsselt
	PublicClientKey []byte `cbor:"4,keyasint"`

	// Speichert die Signatur des Clients ab
	ClientSig []byte `cbor:"5,keyasint"`

	// Speichert den Öffentlichen Sitzungsschlüssel ab
	RandClientPKey []byte `cbor:"6,keyasint"`

	// Speichert die Signatur des Sitzungsschlüssel ab
	RandClientPKeySig []byte `cbor:"7,keyasint"`

	// Speichert die Aktuelle Version des Clients ab
	Version static.RoueXVersion `cbor:"8,keyasint"`

	// Speichert alle Verfügabren Flags ab
	Flags []WSPackageFlag `cbor:"9,keyasint"`
}

// Antwortpaket vom Server
type EncryptedServerHelloPackage struct {
	// Speichert den Öffentlichen Schlüssel des Servers ab
	PublicServerKey []byte `cbor:"10,keyasint"`

	// Speichert den Öffentlichen Schlüssel des Clients ab
	PublicClientKey []byte `cbor:"11,keyasint"`

	// Speichert die Server Signatur ab
	ServerSig []byte `cbor:"12,keyasint"`

	// Speichert den Öffentlichen Schlüssel der Serverseitigen Sitzung
	RandServerPKey []byte `cbor:"13,keyasint"`

	// Speichert die SIgnatur zu dem Serverseitigen Sitzungsschlüssel
	RandServerPKeySig []byte `cbor:"14,keyasint"`

	// Speichert die Aktuelle Version des Clients ab
	Version static.RoueXVersion `cbor:"15,keyasint"`

	// Speichert alle Verfügabren Flags ab
	Flags [16]*WSPackageFlag `cbor:"16,keyasint"`
}

// Stellt das Verschlüsselte Datenpaket dar
type EncryptedTransportPackage struct {
	// Gibt den Pakettypen an
	Type TransportPackageType `cbor:"17,keyasint"`

	// Gibt die Daten des Paketes an
	Data []byte `cbor:"18,keyasint"`
}

func (obj *EncryptedTransportPackage) toBytes() ([]byte, error) {
	data, err := cbor.Marshal(obj, cbor.EncOptions{})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func readEncryptedTransportPackageFromBytes(d_bytes []byte) (*EncryptedTransportPackage, error) {
	var v EncryptedTransportPackage
	if err := cbor.Unmarshal(d_bytes, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Websocket Transportpaket
type WSTransportPaket struct {
	// Stellt die Signatur eines Transportpaketes dar
	Signature []byte `cbor:"21,keyasint"`

	// Stellt die Daten dar
	Data []byte `cbor:"22,keyasint"`
}

func (obj *WSTransportPaket) toBytes() ([]byte, error) {
	data, err := cbor.Marshal(obj, cbor.EncOptions{})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func readWSTransportPaketFromBytes(d_bytes []byte) (*WSTransportPaket, error) {
	var v WSTransportPaket
	if err := cbor.Unmarshal(d_bytes, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Ping Paket
type PingPackage struct {
	// Gibt die ID des aktuellen Ping vorganges an
	PingId []byte `cbor:"23,keyasint"`
}

func (obj *PingPackage) toBytes() ([]byte, error) {
	data, err := cbor.Marshal(obj, cbor.EncOptions{})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func readPingPackageFromBytes(d_bytes []byte) (*PingPackage, error) {
	var v PingPackage
	if err := cbor.Unmarshal(d_bytes, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Pong Paket
type PongPackage struct {
	// Gibt die ID des Pong vorganges an
	PingId []byte `cbor:"24,keyasint"`
}

func (obj *PongPackage) toBytes() ([]byte, error) {
	data, err := cbor.Marshal(obj, cbor.EncOptions{})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func readPongPackageFromBytes(d_bytes []byte) (*PongPackage, error) {
	var v PongPackage
	if err := cbor.Unmarshal(d_bytes, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
