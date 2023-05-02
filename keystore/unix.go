//go:build linux || freebsd || openbsd || netbsd || dragonfly || darwin

package keystore

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/static"
)

func createNewPrivateKe() (*btcec.PublicKey, *btcec.PrivateKey, error) {
	// Es wird ein neuer Privater Schlüssel erstellt
	pr, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	// Der Private Schlüssel wird in Hex umgewadelt
	hexed_priv_key := hex.EncodeToString(pr.Serialize())

	// Der Private Schlüssel wird geschrieben
	file, err := os.Create(static.GetFilePathFor(static.PRIVATE_KEY_FILE))
	if err != nil {
		return nil, nil, err
	}
	_, err = file.WriteString(hexed_priv_key)
	if err != nil {
		return nil, nil, err
	}

	// Die Datei wird geschlossen
	file.Close()

	// Der Öffentliche und Private Schlüssel wird zurückgeben
	fmt.Println("New Private key created to file", static.GetFilePathFor(static.PRIVATE_KEY_FILE))
	return pr.PubKey(), pr, nil
}

func LoadPrivateKeyFromKeyStore() (*btcec.PublicKey, *btcec.PrivateKey, error) {
	// Es wird geprüft ob die Datei vorhanden ist
	fileinfo, err := os.Stat(static.GetFilePathFor(static.PRIVATE_KEY_FILE))
	if os.IsNotExist(err) {
		// Es wird ein neuer Privater und öffentlicher Schlüssel erstellt
		return createNewPrivateKe()
	}

	// Überprüfen, ob der aktuelle Benutzer die Datei lesen darf
	if fileinfo.Mode().Perm()&(1<<2) == 0 {
		return nil, nil, fmt.Errorf("no permissions, key file 1")
	}

	// Es wird geprüft ob die Keydatei vorhanden ist
	content, err := os.ReadFile(static.GetFilePathFor(static.PRIVATE_KEY_FILE))
	if err != nil {
		return nil, nil, fmt.Errorf("error by reading the key file")
	}

	// Es wird versucht die Daten in einen String umzuwandeln
	converted_data := string(content)

	// Es wird geprüft ob es sich um einen 64 Zeichen Langen String handelt
	if len(converted_data) != 64 {
		return nil, nil, fmt.Errorf("invalid private key loaded, panic aborted")
	}

	// Es wird versucht den Privaten Schlüssel einzulesen
	decoded, err := hex.DecodeString(converted_data)
	if err != nil {
		return nil, nil, err
	}

	// Es wird geprüft ob es sich um einen 32 Zeichen Langes Bytes Slice handelt
	if len(decoded) != 32 {
		return nil, nil, fmt.Errorf("invalid private key loaded, panic aborted")
	}

	// Der Private Schlüssel wird erstellt
	privk, pubk := btcec.PrivKeyFromBytes(decoded)

	// Der Private Schlüssel wird zurückgegeben
	fmt.Println("Private key loaded from", static.GetFilePathFor(static.PRIVATE_KEY_FILE))
	return pubk, privk, nil
}
