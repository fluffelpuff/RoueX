package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	ecies "github.com/ecies/go/v2"
	"github.com/fxamacker/cbor"
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/sha3"
)

// Überträgt Verschlüsselte Daten
type _aes_encrypted_result struct {
	Cipher []byte `cbor:"1,keyasint"`
	Nonce  []byte `cbor:"2,keyasint"`
	Sig    []byte `cbor:"3,keyasint"`
}

// Gibt den Verschlüsselungs Algo an
type EncryptionAlgo uint8

// Definiert alle Algos
const (
	CHACHA_2020 = EncryptionAlgo(1)
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

// Erstellt eine Checksume mit einem ECDH Schlüssel
func ComputeChecksumECDH(ecdh_key []byte, data []byte) ([]byte, error) {
	// Erstelle einen SHA-256-Hasher
	hasher := hmac.New(sha256.New, ecdh_key)

	// Schreibe die Nachricht in den Hasher
	hasher.Write(data)

	// Berechne die HMAC
	hmacBytes := hasher.Sum(nil)

	// Die Daten werden zurückgegeben
	return hmacBytes, nil
}

// Verschlüsselt etwas mit AES 256
func EncryptWithChaCha(ecdh_key []byte, data []byte) ([]byte, error) {
	nonce := make([]byte, chacha20.NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		panic(err)
	}

	cipher, err := chacha20.NewUnauthenticatedCipher(ecdh_key, nonce)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, len(data))
	cipher.XORKeyStream(ciphertext, data)

	chsum, err := ComputeChecksumECDH(ecdh_key, ciphertext)
	if err != nil {
		return nil, err
	}

	aes_paket := _aes_encrypted_result{Cipher: ciphertext, Sig: chsum, Nonce: nonce}

	m_data, err := cbor.Marshal(aes_paket, cbor.EncOptions{})
	if err != nil {
		fmt.Println("error:", err)
	}

	return m_data, nil
}

// Verschlüsselt etwas mit AES 256
func DecryptWithChaCha(ecdh_key []byte, data []byte) ([]byte, error) {
	// Erstelle einen CBOR-Decoder
	var aes_paket _aes_encrypted_result
	if err := cbor.Unmarshal(data, &aes_paket); err != nil {
		return nil, err
	}

	// Überprüfe die Checksumme
	chsum, err := ComputeChecksumECDH(ecdh_key, aes_paket.Cipher)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(chsum, aes_paket.Sig) {
		return nil, errors.New("checksum validation failed")
	}

	// Decryption
	cipherDecryption, err := chacha20.NewUnauthenticatedCipher(ecdh_key, aes_paket.Nonce)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(aes_paket.Cipher))
	cipherDecryption.XORKeyStream(decrypted, aes_paket.Cipher)

	// Die Daten werden ohne Fehler zurückgegeben
	return decrypted, nil
}
