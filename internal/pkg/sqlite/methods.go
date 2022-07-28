package sqlite

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/Nunoki/onetimesharer/internal/pkg/randomizer"
)

// DOCME
func (s sqliteStore) ReadSecret(key string) (string, error) {
	hashKey := md5hash(key)

	var content string
	err := s.Database.QueryRowContext(
		s.Context,
		querySelect,
		hashKey,
	).Scan(&content)

	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}

	rawContent, err := s.Crypter.Decrypt(content)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}

	err = s.deleteSecret(hashKey)

	if err != nil {
		return "", fmt.Errorf("failed to delete secret with key %s: %w", key, err)
	}

	return rawContent, nil
}

// DOCME
func (s sqliteStore) SaveSecret(secret string) (string, error) {
	key := randomizer.String(32)
	hashKey := md5hash(key)
	encSecret, err := s.Crypter.Encrypt(secret)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt secret: %w", err)
	}

	_, err = s.Database.ExecContext(
		s.Context,
		queryInsert,
		hashKey,
		encSecret,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store secret: %w", err)
	}

	return key, nil
}

// DOCME
func (s sqliteStore) ValidateSecret(key string) (bool, error) {
	hashKey := md5hash(key)

	err := s.Database.QueryRowContext(
		s.Context,
		querySelect,
		hashKey,
	).Scan(nil)

	if err != nil {
		return false, err
	}

	return err == nil, nil
}

// DOCME
func (s sqliteStore) deleteSecret(hashKey string) error {
	_, err := s.Database.ExecContext(s.Context, queryDelete, hashKey)
	return err
}

// DOCME
func md5hash(input string) (hash string) {
	sum := md5.Sum([]byte(input))
	return hex.EncodeToString(sum[:])
}
