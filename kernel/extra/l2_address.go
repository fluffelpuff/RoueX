package extra

import (
	"bytes"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/utils"
)

type L2Address struct {
	_pkey *btcec.PublicKey
}

func (p *L2Address) Equal(p2 *L2Address) bool {
	return bytes.Equal(p._pkey.SerializeCompressed(), p2._pkey.SerializeCompressed())
}

func (p *L2Address) Hash() uint32 {
	keyHash := hashBytes(p._pkey.SerializeCompressed())
	return keyHash
}

func (p *L2Address) ToString() string {
	return utils.ConvertPublicKeyToAddress(p._pkey)
}

func (p *L2Address) ToByteSlice() []byte {
	return p._pkey.SerializeCompressed()
}

func FromByteSlice(badr []byte) (*L2Address, error) {
	return nil, nil
}

func hashBytes(b []byte) uint32 {
	var hash uint32
	for _, val := range b {
		hash = 31*hash + uint32(val)
	}
	return hash
}
