package utils

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/fluffelpuff/RoueX/static"
)

// Wandelt einen HEX-String in eine Adresse um
func ConvertHexStringToAddress(hxstr string) string {
	// Dekodiere den hexadezimalen String
	decoded, err := hex.DecodeString(hxstr)
	if err != nil {
		panic(err)
	}

	// Kodiere den dekodierten String mit Bech32
	encoded, err := bech32.ConvertBits(decoded, 8, 5, true)
	if err != nil {
		panic(err)
	}

	f, err := bech32.Encode(static.ADDRESS_PREFIX, encoded)
	if err != nil {
		panic(err)
	}

	return f
}

// Wandelt einen Öffentlichen Schlüssel in eine Adresse um
func ConvertPublicKeyToAddress(pubk *btcec.PublicKey) string {
	// Kodiere den dekodierten String mit Bech32
	encoded, err := bech32.ConvertBits(pubk.SerializeCompressed(), 8, 5, true)
	if err != nil {
		panic(err)
	}

	f, err := bech32.Encode(static.ADDRESS_PREFIX, encoded)
	if err != nil {
		panic(err)
	}

	return f
}
