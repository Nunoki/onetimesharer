package crypter

// Crypter is implemented by any value that has an Encrypt() and Decrypt() method. It is used for
// encryption of secrets and their keys before passing it to the storage medium, and corresponding
// decryption when reading it back.
type Crypter interface {
	Decrypt(encrypted string) (string, error)
	Encrypt(plaintext string) (string, error)
}
