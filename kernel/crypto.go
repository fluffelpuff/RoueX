package kernel

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	ecies "github.com/ecies/go/v2"
	"golang.org/x/crypto/sha3"
)

// Wird verwendet um einen SHA3_256 Hash zu erstellen
func ComputeSha3256Hash(data ...[]byte) []byte {
	// Verkette die übergebenen Byte-Slices zu einem einzelnen Byte-Slice
	var combined []byte
	for _, d := range data {
		combined = append(combined, d...)
	}

	// Erstelle einen neuen SHA3-256-Hasher
	hasher := sha3.New256()

	// Schreibe die Daten in den Hasher
	hasher.Write(combined)

	// Berechne den Hashwert und gib ihn zurück
	return hasher.Sum(nil)
}

// Verschlüsselt einen Datensatz mit einem Öffentlichen Schlüssel
func EncryptECIESPublicKey(pkey *btcec.PublicKey, data []byte) ([]byte, error) {
	ecies_pky, err := ecies.NewPublicKeyFromBytes(pkey.SerializeCompressed())
	if err != nil {
		return nil, err
	}

	ciphertext, err := ecies.Encrypt(ecies_pky, data)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

// Ließt einen Privaten Schlüssel ein
func ReadPrivateKeyFromByteSlice(priv_slice []byte) (*btcec.PrivateKey, error) {
	pr, _ := btcec.PrivKeyFromBytes(priv_slice)
	return pr, nil
}

// Ließt einen Öfentlichen Schlüssel aus den Bytes ein
func ReadPublicKeyFromByteSlice(pub_slice []byte) (*btcec.PublicKey, error) {
	return btcec.ParsePubKey(pub_slice)
}

// Wird verwendet um einen Datensatz mit dem eigenen Schlüssel zu entschlüsseln
func DecryptDataWithPrivateKey(priv_key *btcec.PrivateKey, ciphertext []byte) ([]byte, error) {
	dk := ecies.NewPrivateKeyFromBytes(priv_key.Serialize())
	plaintext, err := ecies.Decrypt(dk, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("DecryptDataWithPrivateKey: " + err.Error())
	}
	return plaintext, nil
}

// Erstellt einen Privten Secp256k1 Schlüssel
func GeneratePrivateKey() (*btcec.PrivateKey, error) {
	return btcec.NewPrivateKey()
}

// Signiert einen Hash mit dem Relay Schlüssel
func Sign(priv_key *btcec.PrivateKey, data []byte) ([]byte, error) {
	sig := ecdsa.Sign(priv_key, data)
	return sig.Serialize(), nil
}

// Überprüft die gültigkeit einer Signatur
func VerifyByBytes(public_key *btcec.PublicKey, sig []byte, digest []byte) (bool, error) {
	sigobj, err := ecdsa.ParseDERSignature(sig)
	if err != nil {
		return false, err
	}
	if !sigobj.Verify(digest, public_key) {
		return false, nil
	}
	return true, nil
}
