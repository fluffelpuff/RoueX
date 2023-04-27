package ipoverlay

type EncryptedClientHelloPackage struct {
	PublicServerKey   []byte `cbor:"1,keyasint"`
	PublicClientKey   []byte `cbor:"2,keyasint"`
	ClientSig         []byte `cbor:"3,keyasint"`
	RandClientPKey    []byte `cbor:"4,keyasint"`
	RandClientPKeySig []byte `cbor:"5,keyasint"`
}

type EncryptedServerHelloPackage struct {
	PublicServerKey   []byte `cbor:"6,keyasint"`
	PublicClientKey   []byte `cbor:"7,keyasint"`
	ServerSig         []byte `cbor:"8,keyasint"`
	RandServerPKey    []byte `cbor:"9,keyasint"`
	RandServerPKeySig []byte `cbor:"10,keyasint"`
}
