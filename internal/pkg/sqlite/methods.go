package sqlite

import (
	"fmt"

	"github.com/Nunoki/onetimesharer/internal/pkg/randomizer"
)

// ReadSecret attempts to read the secret corresponding to the passed unencrypted key and returns
// the unencrypted contents of it.
func (s sqliteStore) ReadSecret(key string) (secret string, err error) {
	var encKey string
	encKey, err = s.Crypter.Encrypt(key)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt key when reading: %w", err)
	}

	var encSecret string
	err = s.Database.QueryRowContext(
		s.Context,
		querySelect,
		encKey,
	).Scan(&encSecret)

	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}

	secret, err = s.Crypter.Decrypt(encSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}

	err = s.deleteSecret(encKey)

	if err != nil {
		return "", fmt.Errorf("failed to delete secret with key %s: %w", key, err)
	}

	return secret, nil
}

// SaveSecret will generate a new key, save the contents of the passed secret in an encrypted
// form under the encrypted key, and then return the corresponding unencrypted key.
func (s sqliteStore) SaveSecret(secret string) (key string, err error) {
	key = randomizer.String(32)

	var encKey string
	encKey, err = s.Crypter.Encrypt(key)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt key when saving: %w", err)
	}

	var encSecret string
	encSecret, err = s.Crypter.Encrypt(secret)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt secret: %w", err)
	}

	_, err = s.Database.ExecContext(
		s.Context,
		queryInsert,
		encKey,
		encSecret,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store secret: %w", err)
	}

	return key, nil
}

// ValidateSecret returns whether the requested key exists in the database.
func (s sqliteStore) ValidateSecret(key string) (exists bool, err error) {
	var encKey string
	encKey, err = s.Crypter.Encrypt(key)
	if err != nil {
		return false, fmt.Errorf("failed to encrypt key when validating: %w", err)
	}

	var encSecret string
	err = s.Database.QueryRowContext(
		s.Context,
		querySelect,
		encKey,
	).Scan(&encSecret)

	if err != nil {
		return false, err
	}

	return err == nil, nil
}

// deleteSecret attempts to delete the secret under the passed key from the database. Note that
// this is the only method where the key is passed in its encrypted form (as it is an unexported
// shorthand method).
func (s sqliteStore) deleteSecret(encKey string) error {
	_, err := s.Database.ExecContext(s.Context, queryDelete, encKey)
	return err
}
