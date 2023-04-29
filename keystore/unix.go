//go:build linux || freebsd || openbsd || netbsd || dragonfly || darwin

package keystore

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fluffelpuff/RoueX/static"
)

func LoadPrivateKeyFromKeyStore() (*btcec.PublicKey, *btcec.PrivateKey, error) {
	// Es wird geprüft ob die Datei vorhanden ist
	fileinfo, err := os.Stat(static.GetFilePathFor(static.PRIVATE_KEY_FILE))
	if os.IsNotExist(err) {
		return nil, nil, err
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
	fmt.Println("Privatekey loaded from ", static.GetFilePathFor(static.PRIVATE_KEY_FILE))
	return pubk, privk, nil
}
