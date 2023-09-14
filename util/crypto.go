package util

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// EncryptAES encrypts text with key using AES-256 GCM
func EncryptAES(block cipher.Block, text []byte) ([]byte, error) {

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the plaintext using the GCM cipher
	text = gcm.Seal(nil, nonce, text, nil)

	// Append the nonce to the ciphertext
	text = append(nonce, text...)

	return text, nil
}

// DecryptAES decryptes text with key using AES-256 GCM
func DecryptAES(block cipher.Block, text []byte) ([]byte, error) {

	// Get GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Ensure the ciphertext is of the minimum length
	if len(text) < gcm.NonceSize() {
		return nil, errors.New("ciphertext is too short")
	}

	// Get nonce
	nonce := text[:gcm.NonceSize()]

	plaintext, err := gcm.Open(nil, nonce, text[gcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return key, nil
}

// Hashes using SHA256
func Hash(bytes []byte) []byte {
	hashed := sha256.Sum256(bytes)
	return hashed[:]
}

func CompareHash(h1, h2 []byte) bool {
	return string(h1) == string(h2)
}
