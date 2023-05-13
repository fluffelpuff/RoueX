package utils

import (
	"encoding/base32"
	b32 "encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"

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
	// Das neue Alphabet wird definiert
	customAlphabet := "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	encoder := base32.NewEncoding(customAlphabet)

	// Der Öffentliche Schlüssel wird in
	encoded := encoder.WithPadding(b32.NoPadding).EncodeToString(pubk.SerializeCompressed())

	// Der String wird vorbereitet
	formated_address := fmt.Sprintf("%s1%s", static.ADDRESS_PREFIX, encoded)

	// Die Daten werden zurückgegeben
	return formated_address
}

// Wandelt eine Adresse in einen Öffentlichen Schlüssel um
func ConvertAddressToPublicKey(address_str string) (*btcec.PublicKey, error) {
	// Der String wird gesplittet bei 1
	splited := strings.Split(address_str, "1")

	// Es wird geprüft ob es sich um eine zulässige Adresse handelt
	if splited[0] != static.ADDRESS_PREFIX {
		return nil, fmt.Errorf("ConvertAddressToPublicKey: invalid address")
	}

	// Das neue Alphabet wird definiert
	customAlphabet := "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	encoder := base32.NewEncoding(customAlphabet)

	// Die Adresse wird Dekodiert
	decoded, err := encoder.WithPadding(b32.NoPadding).DecodeString(splited[1])
	if err != nil {
		panic(err)
	}

	// Es wird geprüft ob die Länge des Öffentlichen Schlüssels zulässig ist
	if len(decoded) != 33 {
		return nil, fmt.Errorf("ConvertAddressToPublicKey: invalid public key length: " + string(len(decoded)))
	}

	// Es wird versucht den Öffentlichen Schlüssel einzulesen
	readed_pub_key, err := btcec.ParsePubKey(decoded)
	if err != nil {
		return nil, fmt.Errorf("ConvertAddressToPublicKey: " + err.Error())
	}

	// Der Öffentliche Schlüssel wird zurückgegeben
	return readed_pub_key, nil
}
