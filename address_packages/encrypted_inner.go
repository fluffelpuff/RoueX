package addresspackages

// Stellt die Internen und eigentlichen Verschlüsselten Daten dar
type EncryptedInnerData struct {
	Version  uint32
	Protocol uint8
	Body     []byte
}

// Die Inneren Daten werden in Bytes umgewandelt
func (c *EncryptedInnerData) ToBytes() ([]byte, error) {
	return nil, nil
}
