package utils

import (
	"encoding/base32"
	"encoding/hex"
	"strings"
)

func HexToBase32(h string) (string, error) {
	data, err := hex.DecodeString(h)
	if err != nil {
		return "", err
	}

	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	base32Str := encoder.EncodeToString(data)

	return strings.ToLower(base32Str), nil
}
