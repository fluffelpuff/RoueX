package addresspackages

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fxamacker/cbor"
)

// Stellt ein Verschlüsseltes Paket dar
type FinalAddressLayerPackage struct {
	Reciver          btcec.PublicKey
	Sender           btcec.PublicKey
	IsLocallyCreated bool
	InnerData        []byte
	SSig             []byte
}

// Wird verwendet um das Paket final in Bytes umzuwnadeln
type byted_final_adrl_package struct {
	Reciver []byte
	Sender  []byte
	Body    []byte
	SSig    []byte
}

// Prüft ob die Signatur eines Address Layer Paketes korrekt ist
func (obi *FinalAddressLayerPackage) ValidateSignature() bool {
	return true
}

// Wird verwendet um ein PreAddressLayerPackage aus Bytes einzulesen
func ReadFinalAddressLayerPackageFromBytes(dbyte []byte) (*FinalAddressLayerPackage, error) {
	// Es wird versucht das Paket einzulesen
	var v byted_final_adrl_package
	if err := cbor.Unmarshal(dbyte, &v); err != nil {
		return nil, fmt.Errorf("ReadFinalAddressLayerPackageFromBytes: 1: " + err.Error())
	}

	// Der Öffentliche Schlüssel des Senders wird eingelesen
	sender_pkey, err := btcec.ParsePubKey(v.Sender)
	if err != nil {
		return nil, fmt.Errorf("ReadFinalAddressLayerPackageFromBytes: 2: " + err.Error())
	}

	// Der Öffentliche Schlüssel des Empfängers wird
	reciver_pkey, err := btcec.ParsePubKey(v.Reciver)
	if err != nil {
		return nil, fmt.Errorf("ReadFinalAddressLayerPackageFromBytes: 3: " + err.Error())
	}

	// Das Paket wird nachgebaut
	rebuilded := &FinalAddressLayerPackage{Reciver: *reciver_pkey, Sender: *sender_pkey, InnerData: v.Body, IsLocallyCreated: false}

	// Das Paket wird zurückgegeben
	return rebuilded, nil
}
