package sqlite

import (
	"fmt"

	"github.com/Nunoki/onetimesharer/internal/pkg/randomizer"
)

// DOCME
func (s sqliteStore) ReadSecret(key string) (string, error) {
	encKey, err := s.Crypter.Encrypt(key)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt key when reading: %w", err)
	}

	var content string
	err = s.Database.QueryRowContext(
		s.Context,
		querySelect,
		encKey,
	).Scan(&content)

	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}

	rawContent, err := s.Crypter.Decrypt(content)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}

	err = s.deleteSecret(encKey)

	if err != nil {
		return "", fmt.Errorf("failed to delete secret with key %s: %w", key, err)
	}

	return rawContent, nil
}

// DOCME
func (s sqliteStore) SaveSecret(secret string) (string, error) {
	key := randomizer.String(32)
	encKey, err := s.Crypter.Encrypt(key)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt key when saving: %w", err)
	}

	encSecret, err := s.Crypter.Encrypt(secret)
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

// DOCME
func (s sqliteStore) ValidateSecret(key string) (bool, error) {
	encKey, err := s.Crypter.Encrypt(key)
	if err != nil {
		return false, fmt.Errorf("failed to encrypt key when validating: %w", err)
	}

	var content string
	err = s.Database.QueryRowContext(
		s.Context,
		querySelect,
		encKey,
	).Scan(&content)

	if err != nil {
		return false, err
	}

	return err == nil, nil
}

// DOCME
func (s sqliteStore) deleteSecret(encKey string) error {
	_, err := s.Database.ExecContext(s.Context, queryDelete, encKey)
	return err
}
