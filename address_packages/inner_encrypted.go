package addresspackages

import (
	"fmt"

	"github.com/fxamacker/cbor"
)

// Stellt die Internen und eigentlichen Verschlüsselten Daten dar
type EncryptedInnerData struct {
	Version  uint64 `cbor:"1,keyasint"`
	Protocol uint8  `cbor:"2,keyasint"`
	Body     []byte `cbor:"3,keyasint"`
}

// Die Inneren Daten werden in Bytes umgewandelt
func (c *EncryptedInnerData) ToBytes() ([]byte, error) {
	// Das Paket wird in Bytes umgewandelt
	b_data, err := cbor.Marshal(c, cbor.EncOptions{})
	if err != nil {
		return nil, fmt.Errorf("ToBytes:" + err.Error())
	}

	// Die Daten werden zurückgegeben
	return b_data, nil
}
