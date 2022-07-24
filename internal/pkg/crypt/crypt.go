package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var bytes = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

// Encrypt attempts to encrypt the passed plaintext message
func Encrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
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
func Decrypt(encrypted, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
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
