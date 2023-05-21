package addresspackages

import (
	"encoding/binary"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/utils"
)

// Stellt ein nicht Verschlüsseltes Paket dar
type PreAddressLayerPackage struct {
	Reciver  btcec.PublicKey
	Sender   btcec.PublicKey
	Version  uint64
	Body     []byte
	Protocol uint8
}

// Gibt den Signaturhash aus
func (obj *PreAddressLayerPackage) GetPackageHash() []byte {
	// Das verwendete Protokoll wird in Bytes umgewandelt
	byted_prot := byte(obj.Protocol)

	// Die Version wird umgewandelt
	buf := make([]byte, 8)

	// Schreibe die Zahl ins Byte-Array
	binary.BigEndian.PutUint64(buf, obj.Version)

	// Es wird ein Hash aus dem Paket erstellt
	shash := utils.ComputeSha3256Hash(obj.Reciver.SerializeCompressed(), obj.Sender.SerializeCompressed(), []byte{byted_prot}, buf, obj.Body)

	// Der Signaturhash wird zurückgegeben
	return shash
}
