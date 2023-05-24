package addresspackages

import (
	"fmt"

	"github.com/fxamacker/cbor"
)

// Stellt die Internen und eigentlichen Verschlüsselten Daten dar
type InnerFrame struct {
	Version  uint64 `cbor:"1,keyasint"`
	Protocol uint8  `cbor:"2,keyasint"`
	Data     []byte `cbor:"3,keyasint"`
}

// Die Inneren Daten werden in Bytes umgewandelt
func (c *InnerFrame) ToBytes() ([]byte, error) {
	// Das Paket wird in Bytes umgewandelt
	b_data, err := cbor.Marshal(c, cbor.EncOptions{})
	if err != nil {
		return nil, fmt.Errorf("ToBytes:" + err.Error())
	}

	// Die Daten werden zurückgegeben
	return b_data, nil
}

// Ließt ein Frame aus Bytes ein
func ReadInnerFrameFromBytes(data []byte) (*InnerFrame, error) {
	// Es wird versucht das Paket einzulesen
	var v InnerFrame
	if err := cbor.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("ReadSendableAddressLayerPackageFromBytes: 1: " + err.Error())
	}

	// Das Innerframe wird zurückgegeben
	return &v, nil
}
