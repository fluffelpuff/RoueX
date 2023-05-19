package kernel

import (
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/utils"
	"github.com/fxamacker/cbor"
)

// Stellt ein nicht Verschlüsseltes Paket dar
type PlainAddressLayerPackage struct {
	Reciver          btcec.PublicKey
	Sender           btcec.PublicKey
	IsLocallyCreated bool
	Version          uint32
	Body             []byte
	Protocol         uint8
	SSig             []byte
}

// Stellt die Internen und eigentlichen Verschlüsselten Daten dar
type EncryptedInnerData struct {
	Version  uint32
	Protocol uint8
	Body     []byte
}

// Stellt ein Verschlüsseltes Paket dar
type EncryptedAddressLayerPackage struct {
	Reciver          btcec.PublicKey
	Sender           btcec.PublicKey
	IsLocallyCreated bool
	InnerData        []byte
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

// Die Inneren Daten werden in Bytes umgewandelt
func (c *EncryptedInnerData) ToBytes() ([]byte, error) {
	return nil, nil
}

// Prüft ob die Signatur eines Address Layer Paketes korrekt ist
func (obi *EncryptedAddressLayerPackage) ValidateSignature() bool {
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
	shash := utils.ComputeSha3256Hash(obj.Reciver.SerializeCompressed(), obj.Sender.SerializeCompressed(), []byte{byted_prot}, buf, obj.Body)

	// Der Signaturhash wird zurückgegeben
	return shash
}

// Wird verwendet um ein PlainAddressLayerPackage aus Bytes einzulesen
func ReadEncryptedAddressLayerPackageFromBytes(dbyte []byte) (*EncryptedAddressLayerPackage, error) {
	// Es wird versucht das Paket einzulesen
	var v byted_claim_layer_package
	if err := cbor.Unmarshal(dbyte, &v); err != nil {
		return nil, fmt.Errorf("ReadEncryptedAddressLayerPackageFromBytes: 1: " + err.Error())
	}

	// Der Öffentliche Schlüssel des Senders wird eingelesen
	sender_pkey, err := btcec.ParsePubKey(v.Sender)
	if err != nil {
		return nil, fmt.Errorf("ReadEncryptedAddressLayerPackageFromBytes: 2: " + err.Error())
	}

	// Der Öffentliche Schlüssel des Empfängers wird
	reciver_pkey, err := btcec.ParsePubKey(v.Reciver)
	if err != nil {
		return nil, fmt.Errorf("ReadEncryptedAddressLayerPackageFromBytes: 3: " + err.Error())
	}

	// Das Paket wird nachgebaut
	rebuilded := &EncryptedAddressLayerPackage{Reciver: *reciver_pkey, Sender: *sender_pkey, InnerData: v.Body, IsLocallyCreated: false}

	// Das Paket wird zurückgegeben
	return rebuilded, nil
}
