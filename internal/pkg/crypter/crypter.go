package crypter

type Crypter interface {
	Decrypt(encrypted string) (string, error)
	Encrypt(plaintext string) (string, error)
}
