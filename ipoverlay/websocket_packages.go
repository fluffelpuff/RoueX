package ipoverlay

import (
	"github.com/fxamacker/cbor"
)

// Clientseite P2P Daten
type ClientSideP2PServerSocketData struct {
	ClientP2PServerProtocol ClientServerP2PProtocol `cbor:"17,keyasint"`
	ClientP2PEndPoint       [1024]byte              `cbor:"18,keyasint"`
	ClientP2PEndPintSize    uint16                  `cbor:"19,keyasint"`
}

// Sitzungsinitialisierung
type EncryptedClientHelloPackage struct {
	ClientSideServerData *ClientSideP2PServerSocketData `cbor:"5,keyasint"`
	RandClientPKeySig    []byte                         `cbor:"6,keyasint"`
	HasClientSideP2P     []bool                         `cbor:"7,keyasint"`
	PublicServerKey      []byte                         `cbor:"8,keyasint"`
	PublicClientKey      []byte                         `cbor:"9,keyasint"`
	RandClientPKey       []byte                         `cbor:"10,keyasint"`
	ClientSig            []byte                         `cbor:"11,keyasint"`
	Flags                [256]ProtocolFlag              `cbor:"11,keyasint"`
}

// Antwortpaket vom Server
type EncryptedServerHelloPackage struct {
	PublicServerKey   []byte `cbor:"12,keyasint"`
	PublicClientKey   []byte `cbor:"13,keyasint"`
	ServerSig         []byte `cbor:"14,keyasint"`
	RandServerPKey    []byte `cbor:"15,keyasint"`
	RandServerPKeySig []byte `cbor:"16,keyasint"`
	TryToP2PConnect   bool   `cbor:"17,keyasint"`
}

// Verschlüssltes Transportpaket
type EncryptedTransportPackage struct {
	SourceRelay      []byte               `cbor:"1,keyasint"`
	DestinationRelay []byte               `cbor:"2,keyasint"`
	Type             TransportPackageType `cbor:"3,keyasint"`
	Data             []byte               `cbor:"4,keyasint"`
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
	Signature []byte `cbor:"1,keyasint"`
	Data      []byte `cbor:"2,keyasint"`
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
	PingId []byte `cbor:"1,keyasint"`
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
	PingId []byte `cbor:"1,keyasint"`
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
