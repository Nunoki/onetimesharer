package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Nunoki/onetimesharer/internal/pkg/crypt"
	"github.com/Nunoki/onetimesharer/internal/pkg/randomizer"
)

const filename = "secrets.json"

type Store struct {
	Passphrase string
}

type collection map[string]string

// DOCME
func NewClient(passphrase string) Store {
	if err := verifyFile(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	store := Store{
		Passphrase: passphrase,
	}
	return store
}

// DOCME
func (s Store) ReadSecret(key string) (string, error) {
	secrets, err := readAllSecrets()
	if err != nil {
		return "", err
	}

	eKey, err := crypt.Encrypt(key, s.Passphrase)
	if err != nil {
		return "", err
	}

	eSecret, ok := secrets[eKey]
	if !ok {
		return "", errors.New("not found")
	}

	secret, err := crypt.Decrypt(eSecret, s.Passphrase)
	if err != nil {
		return "", err
	}

	if err := deleteSecret(secrets, eKey); err != nil {
		return "", err
	}

	return secret, nil
}

// DOCME
func (s Store) SaveSecret(secret string) (string, error) {
	key := randomizer.RandStr(32)
	secrets, err := readAllSecrets()
	if err != nil {
		return "", err
	}

	eKey, _ := crypt.Encrypt(key, s.Passphrase)
	eSecret, _ := crypt.Encrypt(secret, s.Passphrase)

	secrets[eKey] = string(eSecret)

	if err := storeSecrets(secrets); err != nil {
		return "", err
	}

	return key, nil
}

// DOCME
func (s Store) ValidateSecret(key string) bool {
	secrets, err := readAllSecrets()
	if err != nil {
		log.Print(err)
		return false
	}

	eKey, err := crypt.Encrypt(key, s.Passphrase)
	if err != nil {
		log.Print(err)
		return false
	}

	_, ok := secrets[eKey]
	return ok
}

// DOCME
func storeSecrets(secrets collection) error {
	jsonData, err := json.Marshal(secrets)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filename, jsonData, os.FileMode(0700)); err != nil {
		return err
	}
	return nil
}

// DOCME
func deleteSecret(secrets collection, key string) error {
	delete(secrets, key)
	if err := storeSecrets(secrets); err != nil {
		return err
	}
	return nil
}

// DOCME
func readAllSecrets() (collection, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	jsonData := collection{}
	err = json.Unmarshal(content, &jsonData)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// verifyFile makes sure the file with the secrets exists, by creating it if it doesn't already.
// If an error occurs with either reading or creating it, it outputs the error and exits the
// program.
func verifyFile() error {
	// TODO: test: https://pkg.go.dev/testing/fstest
	_, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		if err = os.WriteFile(filename, []byte("{}"), os.FileMode(0700)); err != nil {
			return fmt.Errorf("failed to create file: %s", filename)
		}
	}

	if err != nil {
		return fmt.Errorf("could not read file: %s", filename)
	}

	return nil
}
