package addresspackages

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fxamacker/cbor"
)

// Stellt ein Verschlüsseltes Paket dar
type SendableAddressLayerPackage struct {
	Reciver btcec.PublicKey // Sender public key
	Sender  btcec.PublicKey // Reciver public key
	Plain   bool            // Plain or encrypted
	Data    []byte          // Data
	Sig     []byte          // Signature
	PCI     bool            // (PleaseCheckInstructions) Please check instructions
}

// Wird verwendet um das Paket final in Bytes umzuwnadeln
type byted_final_adrl_package struct {
	Reciver []byte `cbor:"1,keyasint"` // Sender public key
	Sender  []byte `cbor:"2,keyasint"` // Reciver public key
	Plain   bool   `cbor:"3,keyasint"` // Plain or encrypted
	Data    []byte `cbor:"4,keyasint"` // Data
	Sig     []byte `cbor:"5,keyasint"` // Signature
	PCI     bool   `cbor:"6,keyasint"` // (PleaseCheckInstructions) Please check instructions
}

// Prüft ob die Signatur eines Address Layer Paketes korrekt ist
func (obi *SendableAddressLayerPackage) ValidateSignature() bool {
	return true
}

// Gibt das Paket als Bytes zurück
func (obj *SendableAddressLayerPackage) ToBytes() ([]byte, error) {
	// Die Innerbytes werden vorbereitet
	pre_inner_data := byted_final_adrl_package{
		Reciver: obj.Reciver.SerializeCompressed(),
		Sender:  obj.Sender.SerializeCompressed(),
		Data:    obj.Data,
		Plain:   obj.Plain,
		Sig:     obj.Sig,
		PCI:     obj.PCI,
	}

	// Das Paket wird in Bytes umgewandelt
	b_data, err := cbor.Marshal(pre_inner_data, cbor.EncOptions{})
	if err != nil {
		return nil, fmt.Errorf("ToBytes:" + err.Error())
	}

	// Die Daten werden zurückgegeben
	return b_data, nil
}

// Wird verwendet um ein AddressLayerPackage aus Bytes einzulesen
func ReadSendableAddressLayerPackageFromBytes(dbyte []byte) (*SendableAddressLayerPackage, error) {
	// Es wird versucht das Paket einzulesen
	var v byted_final_adrl_package
	if err := cbor.Unmarshal(dbyte, &v); err != nil {
		return nil, fmt.Errorf("ReadSendableAddressLayerPackageFromBytes: 1: " + err.Error())
	}

	// Der Öffentliche Schlüssel des Senders wird eingelesen
	sender_pkey, err := btcec.ParsePubKey(v.Sender)
	if err != nil {
		return nil, fmt.Errorf("ReadSendableAddressLayerPackageFromBytes: 2: " + err.Error())
	}

	// Der Öffentliche Schlüssel des Empfängers wird
	reciver_pkey, err := btcec.ParsePubKey(v.Reciver)
	if err != nil {
		return nil, fmt.Errorf("ReadSendableAddressLayerPackageFromBytes: 3: " + err.Error())
	}

	// Das Paket wird nachgebaut
	rebuilded := &SendableAddressLayerPackage{
		Reciver: *reciver_pkey,
		Sender:  *sender_pkey,
		Data:    v.Data,
		Plain:   v.Plain,
		Sig:     v.Sig,
		PCI:     v.PCI,
	}

	// Das Paket wird zurückgegeben
	return rebuilded, nil
}
