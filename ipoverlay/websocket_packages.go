package ipoverlay

import (
	"github.com/fxamacker/cbor"
)

type EncryptedClientHelloPackage struct {
	PublicServerKey   []byte `cbor:"1,keyasint"`
	PublicClientKey   []byte `cbor:"2,keyasint"`
	ClientSig         []byte `cbor:"3,keyasint"`
	RandClientPKey    []byte `cbor:"4,keyasint"`
	RandClientPKeySig []byte `cbor:"5,keyasint"`
}

type EncryptedServerHelloPackage struct {
	PublicServerKey   []byte `cbor:"7,keyasint"`
	PublicClientKey   []byte `cbor:"8,keyasint"`
	ServerSig         []byte `cbor:"9,keyasint"`
	RandServerPKey    []byte `cbor:"10,keyasint"`
	RandServerPKeySig []byte `cbor:"11,keyasint"`
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

func (obj *WSTransportPaket) PreValidate() bool {
	return false
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
