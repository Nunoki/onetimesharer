package aescfb

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var bytes = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

type encrypter struct {
	Key string
}

// New returns a new instance of the AES encrypter in CFB mode, which will use the passed
// encryption key for subsequent operations.
func New(key string) encrypter {
	return encrypter{
		Key: key,
	}
}

// Encrypt attempts to encrypt the passed plaintext message
func (e encrypter) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(e.Key))
	if err != nil {
		return "", err
	}
	plain := []byte(plaintext)
	cfb := cipher.NewCFBEncrypter(block, bytes)
	cipherText := make([]byte, len(plain))
	cfb.XORKeyStream(cipherText, plain)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt tries to decrypt the encrypted message
func (e encrypter) Decrypt(encrypted string) (string, error) {
	block, err := aes.NewCipher([]byte(e.Key))
	if err != nil {
		return "", err
	}
	cipherText, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	cfb := cipher.NewCFBDecrypter(block, bytes)
	plain := make([]byte, len(cipherText))
	cfb.XORKeyStream(plain, cipherText)
	return string(plain), nil
}
