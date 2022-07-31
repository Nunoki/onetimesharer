package filestorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/Nunoki/onetimesharer/internal/pkg/crypter"
	"github.com/Nunoki/onetimesharer/internal/pkg/randomizer"
)

const filename = "secrets.json"

var (
	mutex sync.Mutex
)

type storage struct {
	Crypter crypter.Crypter
}

type collection map[string]string

// New returns an instance of storage with crypter e attached.
func New(e crypter.Crypter) (storage, error) {
	if err := verifyFile(); err != nil {
		return storage{}, fmt.Errorf("failed to verify JSON file for storage: %w", err)
	}

	store := storage{
		Crypter: e,
	}
	return store, nil
}

// ReadSecret tries to read the secret under the passed (unencrypted) key, then decrypt the secret
// and return it in plain text.
func (s storage) ReadSecret(key string) (string, error) {
	secrets, err := readAllSecrets()
	if err != nil {
		return "", err
	}

	eKey, err := s.Crypter.Encrypt(key)
	if err != nil {
		return "", err
	}

	eSecret, ok := secrets[eKey]
	if !ok {
		return "", errors.New("not found")
	}

	secret, err := s.Crypter.Decrypt(eSecret)
	if err != nil {
		return "", err
	}

	if err := deleteSecret(secrets, eKey); err != nil {
		return "", err
	}

	return secret, nil
}

// SaveSecret attempts to generate a new key for the passed secret, and save it in an encrypted
// format, then return its corresponding unencrypted key.
func (s storage) SaveSecret(secret string) (string, error) {
	(&mutex).Lock()
	defer func() {
		(&mutex).Unlock()
	}()

	key := randomizer.String(32)
	secrets, err := readAllSecrets()
	if err != nil {
		return "", err
	}

	eKey, _ := s.Crypter.Encrypt(key)
	eSecret, _ := s.Crypter.Encrypt(secret)

	secrets[eKey] = string(eSecret)

	if err := storeSecrets(secrets); err != nil {
		return "", err
	}

	return key, nil
}

// ValidateSecret returns whether a secret exists under the defined key.
func (s storage) ValidateSecret(key string) (bool, error) {
	secrets, err := readAllSecrets()
	if err != nil {
		return false, err
	}

	eKey, err := s.Crypter.Encrypt(key)
	if err != nil {
		return false, err
	}

	_, ok := secrets[eKey]
	return ok, nil
}

// Close exists to satisfy the server.Storer interface, but doesn't do anything for this
// implementation and cannot fail. Returned error will always be nil.
func (s storage) Close() error {
	return nil
}

// storeSecrets attempts to store all secrets passed in the argument. Passed secrets should already
// be encrypted.
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

// deleteSecret attempts to delete the secret under the passed key from the passed collection of
// secrets. If the passed key already doesn't exist, no error is returned.
func deleteSecret(secrets collection, key string) error {
	delete(secrets, key)
	if err := storeSecrets(secrets); err != nil {
		return err
	}
	return nil
}

// readAllSecrets attempts to read the entire file of secrets, and return them as a collection.
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

// verifyFile makes sure the file with the secrets exists, creating it if necessary.
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
