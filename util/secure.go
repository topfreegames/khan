package util

import (
	"encoding/base64"
	"fmt"

	"github.com/wildlife-studios/crypto"
)

//EncryptData is a func that use wildlife crypto module to cipher the data
// the key must have 32 bytes length
func EncryptData(data string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", &TokenSizeError{Msg: "The key length is different than 32"}
	}

	xChacha := crypto.NewXChacha()
	encrypted, err := xChacha.Encrypt([]byte(data), key)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(encrypted)

	return encoded, nil
}

//DecryptData is a func that use wildlife crypto to decipher the data
// the key must have 32 bytes length
func DecryptData(encodedData string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", &TokenSizeError{Msg: "The key length is different than 32"}
	}

	cipheredData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", err
	}

	xChacha := crypto.NewXChacha()
	data, err := xChacha.Decrypt([]byte(cipheredData), key)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", data), nil
}
