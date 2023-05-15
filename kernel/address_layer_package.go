package kernel

import (
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fxamacker/cbor"
)

// Stellt ein einegelesenes Paket dar
type PlainAddressLayerPackage struct {
	Reciver          btcec.PublicKey
	Sender           btcec.PublicKey
	IsLocallyCreated bool
	Version          uint32
	Body             []byte
	Protocol         uint8
	SSig             []byte
}

// Wird verwendet um das Paket final in Bytes umzuwnadeln
type byted_claim_layer_package struct {
	Reciver  []byte
	Sender   []byte
	Version  uint32
	Protocol uint8
	Body     []byte
	SSig     []byte
}

// Prüft ob die Signatur eines Address Layer Paketes korrekt ist
func (obi *PlainAddressLayerPackage) ValidateSignature() bool {
	return true
}

// Gibt das Paket als Bytes aus
func (obj *PlainAddressLayerPackage) GetPackageBytes() ([]byte, error) {
	// Das Byteframe wird gebaut
	package_byte_preframe := byted_claim_layer_package{Reciver: obj.Reciver.SerializeCompressed(), Sender: obj.Sender.SerializeCompressed(), Version: obj.Version, Protocol: obj.Protocol, Body: obj.Body, SSig: obj.SSig}

	// Das Paket wird in Bytes umgewandelt
	byted_package, err := cbor.Marshal(package_byte_preframe, cbor.EncOptions{})
	if err != nil {
		return nil, err
	}

	// Die Bytes werden zurückgegebn
	return byted_package, nil
}

// Gibt die größe des Paketes ein
func (obj *PlainAddressLayerPackage) GetByteSize() uint32 {
	data, err := obj.GetPackageBytes()
	if err != nil {
		return 0
	}
	return uint32(len(data))
}

// Gibt den Signaturhash aus
func (obj *PlainAddressLayerPackage) GetSignHash() []byte {
	// Das verwendete Protokoll wird in Bytes umgewandelt
	byted_prot := byte(obj.Protocol)

	// Die Version wird umgewandelt
	buf := make([]byte, 4)

	// Schreibe die Zahl ins Byte-Array
	binary.BigEndian.PutUint32(buf, obj.Version)

	// Es wird ein Hash aus dem Paket erstellt
	shash := ComputeSha3256Hash(obj.Reciver.SerializeCompressed(), obj.Sender.SerializeCompressed(), []byte{byted_prot}, buf, obj.Body)

	// Der Signaturhash wird zurückgegeben
	return shash
}

// Wird verwendet um ein PlainAddressLayerPackage aus Bytes einzulesen
func ReadPlainAddressLayerPackageFrameFromBytes(dbyte []byte) (*PlainAddressLayerPackage, error) {
	// Es wird versucht das Paket einzulesen
	var v byted_claim_layer_package
	if err := cbor.Unmarshal(dbyte, &v); err != nil {
		return nil, fmt.Errorf("ReadPlainAddressLayerPackageFrameFromBytes: 1: " + err.Error())
	}

	// Der Öffentliche Schlüssel des Senders wird eingelesen
	sender_pkey, err := btcec.ParsePubKey(v.Sender)
	if err != nil {
		return nil, fmt.Errorf("ReadPlainAddressLayerPackageFrameFromBytes: 2: " + err.Error())
	}

	// Der Öffentliche Schlüssel des Empfängers wird
	reciver_pkey, err := btcec.ParsePubKey(v.Reciver)
	if err != nil {
		return nil, fmt.Errorf("ReadPlainAddressLayerPackageFrameFromBytes: 3: " + err.Error())
	}

	// Das Paket wird nachgebaut
	rebuilded := &PlainAddressLayerPackage{Reciver: *reciver_pkey, Sender: *sender_pkey, Version: v.Version, Body: v.Body, SSig: v.SSig, Protocol: v.Protocol, IsLocallyCreated: false}

	// Das Paket wird zurückgegeben
	return rebuilded, nil
}
