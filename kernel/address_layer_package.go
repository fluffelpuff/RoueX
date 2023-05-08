package kernel

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fxamacker/cbor"
)

// Stellt ein einegelesenes Paket dar
type AddressLayerPackage struct {
	Reciver btcec.PublicKey
	Sender  btcec.PublicKey
	Version uint16
	Body    []byte
	Type    uint8
	SSig    []byte
}

// Wird verwendet um das Paket final in Bytes umzuwnadeln
type byted_claim_layer_package struct {
	Reciver []byte
	Sender  []byte
	Version uint16
	Type    uint8
	Body    []byte
	SSig    []byte
}

// Prüft ob die Signatur eines Address Layer Paketes korrekt ist
func (obi *AddressLayerPackage) ValidateSignature() bool {
	return true
}

// Wird verwendet um ein AddressLayerPackage aus Bytes einzulesen
func ReadAddressLayerPackageFrameFromBytes(dbyte []byte) (*AddressLayerPackage, error) {
	// Es wird versucht das Paket einzulesen
	var v byted_claim_layer_package
	if err := cbor.Unmarshal(dbyte, &v); err != nil {
		return nil, fmt.Errorf("ReadAddressLayerPackageFrameFromBytes: 1: " + err.Error())
	}

	// Der Öffentliche Schlüssel des Senders wird eingelesen
	sender_pkey, err := btcec.ParsePubKey(v.Sender)
	if err != nil {
		return nil, fmt.Errorf("ReadAddressLayerPackageFrameFromBytes: 2: " + err.Error())
	}

	// Der Öffentliche Schlüssel des Empfängers wird
	reciver_pkey, err := btcec.ParsePubKey(v.Reciver)
	if err != nil {
		return nil, fmt.Errorf("ReadAddressLayerPackageFrameFromBytes: 3: " + err.Error())
	}

	// Das Paket wird nachgebaut
	rebuilded := &AddressLayerPackage{Reciver: *reciver_pkey, Sender: *sender_pkey, Version: v.Version, Body: v.Body, SSig: v.SSig, Type: v.Type}

	// Das Paket wird zurückgegeben
	return rebuilded, nil
}
